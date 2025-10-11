package controllers

import (
	"testing"

	obs "github.com/seventh-horizon/observatory-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestComputeFrontier(t *testing.T) {
	g := NewWithT(t)
	run := &obs.ObservatoryRun{
		Spec: obs.ObservatoryRunSpec{
			Workflow: obs.WorkflowSpec{
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
			Workflow: obs.WorkflowSpec{
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
