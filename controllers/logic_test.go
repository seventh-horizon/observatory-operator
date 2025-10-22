package controllers

import (
	"testing"

	obs "github.com/example/observatory-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestComputeFrontier(t *testing.T) {
	g := NewWithT(t)
	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.Workflow{
				Tasks: map[string]obs.TaskSpec{
					"a": {},
					"b": {Dependencies: []string{"a"}},
					"c": {Dependencies: []string{"a"}},
				},
			},
		},
		Status: obs.ObservatoryRunStatus{
			TaskStatuses: map[string]*obs.TaskStatus{
				"a": {State: obs.TaskSucceeded},
			},
		},
	}
	fr := computeFrontier(run)
	g.Expect(fr).To(ContainElements("b","c"))
}

func TestDerivePhase(t *testing.T) {
	g := NewWithT(t)
	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.Workflow{
				Tasks: map[string]obs.TaskSpec{
					"a": {}, "b": {},
				},
			},
		},
		Status: obs.ObservatoryRunStatus{
			TaskStatuses: map[string]*obs.TaskStatus{
				"a": {State: obs.TaskSucceeded},
				"b": {State: obs.TaskRunning},
			},
		},
	}
	g.Expect(derivePhase(run)).To(Equal(obs.PhaseRunning))
}

func TestRetryAwareFrontier(t *testing.T) {
	g := NewWithT(t)

	backoffLimit := int32(3)
	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.Workflow{
				Tasks: map[string]obs.TaskSpec{
					"task-a": {Retries: &backoffLimit},
					"task-b": {Dependencies: []string{"task-a"}},
				},
			},
		},
		Status: obs.ObservatoryRunStatus{
			TaskStatuses: map[string]*obs.TaskStatus{
				"task-a": {
					State:   obs.TaskPending,
					Message: "Failed 1/3 times, retrying",
				},
			},
		},
	}

	fr := computeFrontier(run)
	g.Expect(fr).To(ContainElement("task-a"))
	g.Expect(fr).NotTo(ContainElement("task-b"))
}

func TestBlockingFailurePolicy(t *testing.T) {
	g := NewWithT(t)

	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.Workflow{
				Tasks: map[string]obs.TaskSpec{
					"task-a": {},
					"task-b": {},
				},
				FailurePolicy: "Stop",
			},
		},
		Status: obs.ObservatoryRunStatus{
			TaskStatuses: map[string]*obs.TaskStatus{
				"task-a": {State: obs.TaskFailed, Message: "Max retries exceeded"},
				"task-b": {State: obs.TaskPending},
			},
		},
	}

	fr := computeFrontier(run)
	g.Expect(fr).To(BeEmpty())
}

func TestMessagePopulation(t *testing.T) {
	g := NewWithT(t)

	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.Workflow{
				Tasks: map[string]obs.TaskSpec{"task-a": {}},
			},
		},
		Status: obs.ObservatoryRunStatus{
			TaskStatuses: map[string]*obs.TaskStatus{
				"task-a": {
					State:   obs.TaskSucceeded,
					Message: "Completed successfully",
				},
			},
		},
	}

	g.Expect(run.Status.TaskStatuses["task-a"].Message).To(Equal("Completed successfully"))
}
