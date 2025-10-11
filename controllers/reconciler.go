package controllers

import (
	"context"
	"fmt"
	"sort"

	observatoryv1alpha1 "github.com/seventh-horizon/observatory-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelRun = "obs.seventh/run"
)

type ObservatoryRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ObservatoryRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&observatoryv1alpha1.ObservatoryRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

func (r *ObservatoryRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var run observatoryv1alpha1.ObservatoryRun
	if err := r.Get(ctx, req.NamespacedName, &run); err != nil {
		if apierrors.IsNotFound(err) { return ctrl.Result{}, nil }
		return ctrl.Result{}, err
	}
	if run.Status.TaskStatuses == nil {
		run.Status.TaskStatuses = map[string]*observatoryv1alpha1.TaskStatus{}
	}

	// Sync status from Jobs
	if err := r.collectJobStatuses(ctx, &run); err != nil {
		return ctrl.Result{}, err
	}

	// Compute frontier (ready tasks)
	frontier := computeFrontier(&run)

	// Launch Jobs for frontier
	for _, t := range frontier {
		if err := r.ensureJobForTask(ctx, &run, t); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Derive phase
	old := run.Status.Phase
	run.Status.Phase = derivePhase(&run)
	if old != run.Status.Phase {
		logger.Info("Phase change", "old", old, "new", run.Status.Phase)
	}

	if err := r.Status().Update(ctx, &run); err != nil {
		return ctrl.Result{}, err
	}

	// Requeue while not terminal
	if run.Status.Phase == observatoryv1alpha1.PhaseSucceeded || run.Status.Phase == observatoryv1alpha1.PhaseFailed {
		return ctrl.Result{}, nil
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *ObservatoryRunReconciler) ensureJobForTask(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun, task string) error {
	jobName := fmt.Sprintf("%s-%s", run.Name, task)
	var existing batchv1.Job
	if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: run.Namespace}, &existing); err == nil {
		return nil
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	ts := run.Spec.Workflow.Tasks[task]
	image := ts.Image
	if image == "" {
		image = "busybox:1.36"
	}
	cmd := []string{"/bin/sh", "-c"}
	arg := ts.Command
	if arg == "" {
		arg = "echo hello from " + task + " && sleep 3"
	}
	backoff := int32(0)
	if ts.Retries != nil {
		backoff = *ts.Retries
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: run.Namespace,
			Labels:    map[string]string{labelRun: run.Name},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoff,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:  "task",
						Image: image,
						Command: cmd,
						Args: []string{arg},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{Command: []string{"sh","-c","true"}},
							},
							InitialDelaySeconds: 1,
							PeriodSeconds: 5,
							FailureThreshold: 3,
							TimeoutSeconds: 2,
							SuccessThreshold: 1,
						},
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(run, job, r.Scheme); err != nil { return err }
	if err := r.Create(ctx, job); err != nil {
		return NewJobCreationError(task, err)
	}
	// update status record with job name
	if run.Status.TaskStatuses[task] == nil {
		run.Status.TaskStatuses[task] = &observatoryv1alpha1.TaskStatus{}
	}
	run.Status.TaskStatuses[task].JobName = jobName
	return nil
}

func (r *ObservatoryRunReconciler) collectJobStatuses(ctx context.Context, run *observatoryv1alpha1.ObservatoryRun) error {
	var jl batchv1.JobList
	if err := r.List(ctx, &jl, client.InNamespace(run.Namespace), client.MatchingLabels{labelRun: run.Name}); err != nil {
		return err
	}
	// reset states we will recompute
	for name := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[name]
		if st == nil {
			run.Status.TaskStatuses[name] = &observatoryv1alpha1.TaskStatus{State: observatoryv1alpha1.TaskPending}
		}
	}
	for _, j := range jl.Items {
		// task name is job name minus "<runname>-"
		prefix := run.Name + "-"
		task := j.Name
		if len(task) > len(prefix) && task[:len(prefix)] == prefix {
			task = task[len(prefix):]
		}
		st := run.Status.TaskStatuses[task]
		if st == nil {
			st = &observatoryv1alpha1.TaskStatus{}
			run.Status.TaskStatuses[task] = st
		}
		st.JobName = j.Name
		// derive state
		switch {
		case j.Status.Succeeded > 0:
			st.State = observatoryv1alpha1.TaskSucceeded
		case j.Status.Failed > 0 && (j.Spec.BackoffLimit != nil && j.Status.Failed >= *j.Spec.BackoffLimit):
			st.State = observatoryv1alpha1.TaskFailed
		case j.Status.Active > 0:
			st.State = observatoryv1alpha1.TaskRunning
		default:
			if st.State == "" { st.State = observatoryv1alpha1.TaskPending }
		}
	}
	return nil
}

func computeFrontier(run *observatoryv1alpha1.ObservatoryRun) []string {
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
				depsOK = false
				break
			}
		}
		if depsOK {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)
	return ready
}

func derivePhase(run *observatoryv1alpha1.ObservatoryRun) observatoryv1alpha1.Phase {
	total := len(run.Spec.Workflow.Tasks)
	if total == 0 { return observatoryv1alpha1.PhasePending }
	succ := 0
	fail := 0
	runAny := 0
	for name := range run.Spec.Workflow.Tasks {
		st := run.Status.TaskStatuses[name]
		if st == nil || st.State == observatoryv1alpha1.TaskPending { continue }
		if st.State == observatoryv1alpha1.TaskRunning { runAny++ }
		if st.State == observatoryv1alpha1.TaskSucceeded { succ++ }
		if st.State == observatoryv1alpha1.TaskFailed { fail++ }
	}
	switch {
	case succ == total:
		return observatoryv1alpha1.PhaseSucceeded
	case fail > 0:
		return observatoryv1alpha1.PhaseFailed
	case runAny > 0 || succ > 0:
		return observatoryv1alpha1.PhaseRunning
	default:
		return observatoryv1alpha1.PhasePending
	}
}
