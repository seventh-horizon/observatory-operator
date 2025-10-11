package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TaskSpec struct {
	Type         string   `json:"type,omitempty"`
	Image        string   `json:"image,omitempty"`
	Command      string   `json:"command,omitempty"`
	Args         []string `json:"args,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Retries      *int32   `json:"retries,omitempty"`
}

type WorkflowSpec struct {
	Tasks map[string]TaskSpec `json:"tasks,omitempty"`
}

type ResourcesSpec struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

type ObservabilitySpec struct {
	OTel *OTelSpec `json:"otel,omitempty"`
}

type OTelSpec struct {
	Enabled    bool              `json:"enabled,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type ObservatoryRunSpec struct {
	Project       string           `json:"project,omitempty"`
	Workflow      WorkflowSpec     `json:"workflow"`
	Resources     *ResourcesSpec   `json:"resources,omitempty"`
	Observability *ObservabilitySpec `json:"observability,omitempty"`
}

type TaskState string
const (
	TaskPending   TaskState = "Pending"
	TaskRunning   TaskState = "Running"
	TaskSucceeded TaskState = "Succeeded"
	TaskFailed    TaskState = "Failed"
)

type TaskStatus struct {
	State   TaskState `json:"state,omitempty"`
	JobName string    `json:"jobName,omitempty"`
	Message string    `json:"message,omitempty"`
}

type Phase string
const (
	PhasePending   Phase = "Pending"
	PhaseRunning   Phase = "Running"
	PhaseSucceeded Phase = "Succeeded"
	PhaseFailed    Phase = "Failed"
)

type ObservatoryRunStatus struct {
	Phase        Phase                 `json:"phase,omitempty"`
	TaskStatuses map[string]*TaskStatus `json:"taskStatuses,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ObservatoryRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObservatoryRunSpec   `json:"spec,omitempty"`
	Status ObservatoryRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ObservatoryRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObservatoryRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObservatoryRun{}, &ObservatoryRunList{})
}
