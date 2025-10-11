package controllers

import (
	"context"
	"fmt"
	"strings"

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
	"sigs.k8s.io/controller-runtime/pkg/record"
)

const (
	labelRun = "obs.seventh/run"
	finalizerName = "observatory.seventh-horizon.io/finalizer"
)

type ObservatoryRunReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      ctrl.Logger
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

	if err := r.collectJobStatuses(ctx, &run); err != nil {
		return ctrl.Result{}, err
	}

	frontier := r.computeFrontier(&run)
	for _, t := range frontier {
		if err := r.ensureJob(ctx, &run, t); err != nil {
			return ctrl.Result{}, err
		}
	}

	run.Status.Phase = r.derivePhase(&run)
	if err := r.Status().Update(ctx, &run); err != nil {
		logger.Error(err, "status update failed")
		return ctrl.Result{}, err
	}

	if run.Status.Phase == observatoryv1alpha1.PhaseSucceeded || run.Status.Phase == observatoryv1alpha1.PhaseFailed {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *ObservatoryRunReconciler) collectJobStatuses(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun) error {
	var jobs batchv1.JobList
	if err := r.List(ctx, &jobs, client.InNamespace(run.Namespace), client.MatchingLabels{labelRun: run.Name}); err != nil {
		return err
	}
	for name := range run.Spec.Workflow.Tasks {
		js := &observatoryv1alpha1.TaskStatus{State: observatoryv1alpha1.TaskPending}
		run.Status.TaskStatuses[name] = js
	}
	for _, j := range jobs.Items {
		name := strings.TrimPrefix(j.Name, run.Name+"-")
		st := run.Status.TaskStatuses[name]
		if st == nil { st = &observatoryv1alpha1.TaskStatus{}; run.Status.TaskStatuses[name]=st }
		st.JobName = j.Name
		if j.Status.Succeeded > 0 { st.State = observatoryv1alpha1.TaskSucceeded
		} else if j.Status.Failed > 0 { st.State = observatoryv1alpha1.TaskFailed
		} else if j.Status.Active > 0 { st.State = observatoryv1alpha1.TaskRunning
		} else { st.State = observatoryv1alpha1.TaskPending }
	}
	return nil
}

func (r *ObservatoryRunReconciler) computeFrontier(run *observatoryv1alpha1.ObservatoryRun) []string {
	ready := []string{}
	for name, spec := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[name]
		if st != nil && (st.State == observatoryv1alpha1.TaskRunning || st.State == observatoryv1alpha1.TaskSucceeded) {
			continue
		}
		depsOK := true
		for _, d := range spec.Dependencies {
			dst := run.Status.TaskStatuses[d]
			if dst == nil || dst.State != observatoryv1alpha1.TaskSucceeded {
				depsOK = false; break
			}
		}
		if depsOK { ready = append(ready, name) }
	}
	return ready
}

func (r *ObservatoryRunReconciler) derivePhase(run *observatoryv1alpha1.ObservatoryRun) observatoryv1alpha1.Phase {
	succeeded, failed, total := 0, 0, len(run.Spec.Workflow.Tasks)
	for n := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[n]
		if st == nil { continue }
		if st.State == observatoryv1alpha1.TaskFailed { failed++ }
		if st.State == observatoryv1alpha1.TaskSucceeded { succeeded++ }
	}
	if failed > 0 { return observatoryv1alpha1.PhaseFailed }
	if succeeded == total && total > 0 { return observatoryv1alpha1.PhaseSucceeded }
	if succeeded > 0 { return observatoryv1alpha1.PhaseRunning }
	return observatoryv1alpha1.PhasePending
}

func (r *ObservatoryRunReconciler) ensureJob(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun, task string) error {
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
	r.Recorder.Event(run, "Normal", "JobCreated", fmt.Sprintf("Created Job %s", jobName))
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
