package v1alpha1

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var observatoryrunlog = logf.Log.WithName("observatoryrun-resource")

func (r *ObservatoryRun) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(r).Complete()
}

// +kubebuilder:webhook:path=/validate-observatory-seventh-horizon-io-v1alpha1-observatoryrun,mutating=false,failurePolicy=fail,sideEffects=None,groups=observatory.seventh-horizon.io,resources=observatoryruns,verbs=create;update,versions=v1alpha1,name=vobservatoryrun.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ObservatoryRun{}

func (r *ObservatoryRun) ValidateCreate() (admission.Warnings, error) {
	observatoryrunlog.Info("validate create", "name", r.Name)
	return r.validate()
}

func (r *ObservatoryRun) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	observatoryrunlog.Info("validate update", "name", r.Name)
	return r.validate()
}

func (r *ObservatoryRun) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *ObservatoryRun) validate() (admission.Warnings, error) {
	var errs []string
	var warns admission.Warnings

	if len(r.Spec.Workflow.Tasks) == 0 {
		errs = append(errs, "workflow must have at least one task")
	}

	for name, spec := range r.Spec.Workflow.Tasks {
		if err := validateTaskName(name); err != nil {
			errs = append(errs, fmt.Sprintf("task '%s': %v", name, err))
		}
		for _, dep := range spec.Dependencies {
			if _, ok := r.Spec.Workflow.Tasks[dep]; !ok {
				errs = append(errs, fmt.Sprintf("task '%s' depends on non-existent task '%s'", name, dep))
			}
			if dep == name {
				errs = append(errs, fmt.Sprintf("task '%s' cannot depend on itself", name))
			}
		}
		if spec.Retries != nil && *spec.Retries < 0 {
			errs = append(errs, fmt.Sprintf("task '%s': retries cannot be negative", name))
		}
		if spec.Retries != nil && *spec.Retries > 10 {
			warns = append(warns, fmt.Sprintf("task '%s' has high retry count (%d)", name, *spec.Retries))
		}
	}
	if err := validateNoCycles(r.Spec.Workflow.Tasks); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return warns, fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return warns, nil
}

func validateTaskName(name string) error {
	if name == "" { return fmt.Errorf("task name cannot be empty") }
	if len(name) > 63 { return fmt.Errorf("task name too long (max 63)") }
	if name[0] == '-' || name[0] == '_' || name[len(name)-1] == '-' || name[len(name)-1] == '_' {
		return fmt.Errorf("task name cannot start or end with '-' or '_'")
	}
	for i, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return fmt.Errorf("invalid character '%c' at position %d", r, i)
	}
	return nil
}

func validateNoCycles(tasks map[string]TaskSpec) error {
	graph := map[string][]string{}
	for n, t := range tasks {
		graph[n] = append([]string{}, t.Dependencies...)
	}
	visited := map[string]bool{}
	rec := map[string]bool{}
	var dfs func(string) bool
	dfs = func(n string) bool {
		visited[n] = true
		rec[n] = true
		for _, d := range graph[n] {
			if !visited[d] {
				if dfs(d) { return true }
			} else if rec[d] { return true }
		}
		rec[n] = false
		return false
	}
	for n := range graph {
		if !visited[n] {
			if dfs(n) { return fmt.Errorf("circular dependency detected") }
		}
	}
	return nil
}
