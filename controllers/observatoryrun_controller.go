package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	observatoryv1alpha1 "github.com/example/observatory-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	labelRun = "obs.seventh/run"
	finalizerName = "observatory.seventh-horizon.io/finalizer"
)

type ObservatoryRunReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
}

func (r *ObservatoryRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var run observatoryv1alpha1.ObservatoryRun
	if err := r.Get(ctx, req.NamespacedName, &run); err != nil {
		if apierrors.IsNotFound(err) { return ctrl.Result{}, nil }
		return ctrl.Result{}, err
	}

	if run.Status.TaskStatuses == nil { run.Status.TaskStatuses = map[string]*observatoryv1alpha1.TaskStatus{} }

	if !run.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &run)
	}

	if !controllerutil.ContainsFinalizer(&run, finalizerName) {
		controllerutil.AddFinalizer(&run, finalizerName)
		if err := r.Update(ctx, &run); err != nil { return ctrl.Result{}, err }
	}

	// Create a deep copy to patch from (must be BEFORE any status mutations)
	orig := run.DeepCopy()

	if err := r.collectJobStatuses(ctx, &run); err != nil {
		return ctrl.Result{}, err
	}

	frontier := r.computeFrontier(&run)
	for _, t := range frontier {
		if err := r.ensureJob(ctx, &run, t); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update the phase or other fields in status
	run.Status.Phase = r.derivePhase(&run)

	// Apply a merge patch to avoid resourceVersion conflicts
	if err := r.Status().Patch(ctx, &run, client.MergeFrom(orig)); err != nil {
		if apierrors.IsConflict(err) {
			// Retry later; object was modified concurrently
			return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
		logger.Error(err, "status patch failed")
		return ctrl.Result{}, err
	}

	if run.Status.Phase == observatoryv1alpha1.PhaseSucceeded || run.Status.Phase == observatoryv1alpha1.PhaseFailed {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *ObservatoryRunReconciler) collectJobStatuses(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun) error {
	// Defensive: ensure the status map always exists (guards against nil map on fresh objects or concurrent updates)
	if run.Status.TaskStatuses == nil {
		run.Status.TaskStatuses = map[string]*observatoryv1alpha1.TaskStatus{}
	}
	var jobs batchv1.JobList
	if err := r.List(ctx, &jobs, client.InNamespace(run.Namespace), client.MatchingLabels{labelRun: run.Name}); err != nil {
		return err
	}
	// Initialize all declared tasks to Pending unless overwritten by observed Jobs
	for name := range run.Spec.Workflow.Tasks {
		if run.Status.TaskStatuses[name] == nil {
			run.Status.TaskStatuses[name] = &observatoryv1alpha1.TaskStatus{State: observatoryv1alpha1.TaskPending}
		}
	}
	for _, j := range jobs.Items {
		name := strings.TrimPrefix(j.Name, run.Name+"-")
		st := run.Status.TaskStatuses[name]
		if st == nil {
			st = &observatoryv1alpha1.TaskStatus{}
			run.Status.TaskStatuses[name] = st
		}
		st.JobName = j.Name

		// Derive human-friendly status message helpers
		started := ""
		if j.Status.StartTime != nil {
			started = j.Status.StartTime.Time.UTC().Format(time.RFC3339)
		}

		switch {
		case j.Status.Succeeded > 0:
			st.State = observatoryv1alpha1.TaskSucceeded
			if j.Status.CompletionTime != nil && j.Status.StartTime != nil {
				dur := j.Status.CompletionTime.Sub(j.Status.StartTime.Time).Round(time.Second)
				st.Message = fmt.Sprintf("Completed successfully in %s", dur)
			} else {
				st.Message = "Completed successfully"
			}

		case j.Status.Failed > 0:
			if j.Spec.BackoffLimit != nil && j.Status.Failed < *j.Spec.BackoffLimit {
				// Retry still allowed: present as Pending so computeFrontier can pick it back up
				st.State = observatoryv1alpha1.TaskPending
				st.Message = fmt.Sprintf("Failed %d/%d times, retrying", j.Status.Failed, *j.Spec.BackoffLimit)
			} else {
				st.State = observatoryv1alpha1.TaskFailed
				if j.Spec.BackoffLimit != nil {
					st.Message = fmt.Sprintf("Failed after %d/%d attempts", j.Status.Failed, *j.Spec.BackoffLimit)
				} else {
					st.Message = fmt.Sprintf("Failed after %d attempts", j.Status.Failed)
				}
			}

		case j.Status.Active > 0:
			st.State = observatoryv1alpha1.TaskRunning
			if started != "" {
				st.Message = fmt.Sprintf("Started at %s", started)
			} else {
				st.Message = "Running"
			}

		default:
			if st.State == "" {
				st.State = observatoryv1alpha1.TaskPending
			}
			if st.Message == "" {
				st.Message = "Waiting for dependencies"
			}
		}
	}
	return nil
}

func (r *ObservatoryRunReconciler) computeFrontier(run *observatoryv1alpha1.ObservatoryRun) []string {
	// If FailurePolicy is Stop and any task has failed, block scheduling new tasks
	if strings.EqualFold(run.Spec.Workflow.FailurePolicy, "Stop") {
		for _, st := range run.Status.TaskStatuses {
			if st != nil && st.State == observatoryv1alpha1.TaskFailed {
				return []string{}
			}
		}
	}

	ready := []string{}
	for name, spec := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[name]
		// Skip if already running or completed
		if st != nil && (st.State == observatoryv1alpha1.TaskRunning || st.State == observatoryv1alpha1.TaskSucceeded) {
			continue
		}
		// All dependencies must have succeeded
		depsOK := true
		for _, d := range spec.Dependencies {
			dst := run.Status.TaskStatuses[d]
			if dst == nil || dst.State != observatoryv1alpha1.TaskSucceeded {
				depsOK = false
				break
			}
		}
		if depsOK {
			ready = append(ready, name)
		}
	}
	return ready
}

func (r *ObservatoryRunReconciler) derivePhase(run *observatoryv1alpha1.ObservatoryRun) observatoryv1alpha1.Phase {
	// Aggregate task states
	succeeded := 0
	failed := 0
	running := 0
	pending := 0
	total := len(run.Spec.Workflow.Tasks)

	for name := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[name]
		if st == nil {
			pending++
			continue
		}
		switch st.State {
		case observatoryv1alpha1.TaskSucceeded:
			succeeded++
		case observatoryv1alpha1.TaskFailed:
			failed++
		case observatoryv1alpha1.TaskRunning:
			running++
		default:
			pending++
		}
	}

	// Any failure flips the whole run to Failed (fail-fast semantics for the run overall)
	if failed > 0 {
		return observatoryv1alpha1.PhaseFailed
	}

	// All tasks succeeded
	if total > 0 && succeeded == total {
		return observatoryv1alpha1.PhaseSucceeded
	}

	// If at least one task has started (Running or Succeeded) or there are active Jobs,
	// the run is considered Running. This covers the case where some tasks are still
	// Pending due to dependencies while others have started.
	if running > 0 || succeeded > 0 {
		return observatoryv1alpha1.PhaseRunning
	}

	// Otherwise nothing has started yet
	return observatoryv1alpha1.PhasePending
}

func (r *ObservatoryRunReconciler) ensureJob(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun, task string) error {
	logger := log.FromContext(ctx)
	jobName := fmt.Sprintf("%s-%s", run.Name, task)
	var existing batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: run.Namespace}, &existing); err == nil {
		return nil
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	spec := run.Spec.Workflow.Tasks[task]
	image := spec.Image
	if image == "" { image = "busybox:1.36" }

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName, Namespace: run.Namespace,
			Labels: map[string]string{labelRun: run.Name},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: spec.Retries,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:  "task",
						Image: image,
						Command: commandFor(spec),
					}},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(run, job, r.Scheme); err != nil { return err }
	if err := r.Create(ctx, job); err != nil { return err }
	logger.Info("Created Job", "job", jobName, "task", task)
	return nil
}

func commandFor(spec observatoryv1alpha1.TaskSpec) []string {
	if spec.Command != "" {
		return []string{"/bin/sh", "-lc", spec.Command}
	}
	if len(spec.Args) > 0 { return spec.Args }
	return []string{"/bin/sh", "-lc", "echo running && sleep 2 && echo done"}
}

func (r *ObservatoryRunReconciler) handleDeletion(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun) (ctrl.Result, error) {
	controllerutil.RemoveFinalizer(run, finalizerName)
	if err := r.Update(ctx, run); err != nil { return ctrl.Result{}, err }
	return ctrl.Result{}, nil
}

func (r *ObservatoryRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&observatoryv1alpha1.ObservatoryRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
