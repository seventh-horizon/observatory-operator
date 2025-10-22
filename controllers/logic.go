package controllers

import (
	"slices"

	obs "github.com/example/observatory-operator/api/v1alpha1"
)

// computeFrontier returns task names whose dependencies have succeeded
// and that are still pending.
func computeFrontier(run *obs.ObservatoryRun) []string {
	// ✨ If failurePolicy is "Stop", block launching anything after a failure
	if run != nil && run.Spec.Workflow.FailurePolicy == "Stop" {
		for _, st := range run.Status.TaskStatuses {
			if st != nil && st.State == obs.TaskFailed {
				return []string{}
			}
		}
	}

	if run == nil || run.Spec.Workflow.Tasks == nil {
		return nil
	}

	var frontier []string
	for name, spec := range run.Spec.Workflow.Tasks {
		status, exists := run.Status.TaskStatuses[name]
		// Skip if task already started or completed
		if exists && status.State != obs.TaskPending {
			continue
		}
		// Check dependencies
		depsOK := true
		for _, dep := range spec.Dependencies {
			depStatus, ok := run.Status.TaskStatuses[dep]
			if !ok || depStatus.State != obs.TaskSucceeded {
				depsOK = false
				break
			}
		}
		if depsOK {
			frontier = append(frontier, name)
		}
	}
	slices.Sort(frontier)
	return frontier
}

// derivePhase summarizes the overall workflow state from individual tasks.
func derivePhase(run *obs.ObservatoryRun) obs.Phase {
	if run == nil {
		return obs.PhasePending
	}

	total := len(run.Spec.Workflow.Tasks)
	if total == 0 {
		return obs.PhasePending
	}

	hasRunning := false
	hasFailed := false
	done := 0

	for _, status := range run.Status.TaskStatuses {
		switch status.State {
		case obs.TaskFailed:
			hasFailed = true
			done++
		case obs.TaskSucceeded:
			done++
		case obs.TaskRunning:
			hasRunning = true
		}
	}

	switch {
	case hasFailed:
		return obs.PhaseFailed
	case done == total:
		return obs.PhaseSucceeded
	case hasRunning || done > 0:
		return obs.PhaseRunning
	default:
		return obs.PhasePending
	}
}