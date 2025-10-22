/*
Copyright 2024 The Observatory Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskSpec defines a single task in the workflow DAG
type TaskSpec struct {
	// Name is the unique identifier for this task
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Image is the container image to run for this task
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// Command to run in the container
	// +optional
	Command []string `json:"command,omitempty"`

	// Args are the arguments to pass to the command
	// +optional
	Args []string `json:"args,omitempty"`

	// Dependencies lists task names that must complete before this task runs
	// +optional
	Dependencies []string `json:"dependencies,omitempty"`

	// Env are environment variables to set in the container
	// +optional
	Env []EnvVar `json:"env,omitempty"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	// Name of the environment variable
	Name string `json:"name"`

	// Value of the environment variable
	// +optional
	Value string `json:"value,omitempty"`
}

// ObservatorySpec defines the desired state of Observatory
type ObservatorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Tasks defines the list of tasks in the workflow DAG
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Tasks []TaskSpec `json:"tasks"`

	// Schedule defines when the workflow should run (optional, cron format)
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Suspend indicates if the workflow execution should be suspended
	// +optional
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

	// MaxConcurrent defines the maximum number of concurrent task executions
	// +optional
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	MaxConcurrent int32 `json:"maxConcurrent,omitempty"`
}

// TaskStatus represents the status of a single task
type TaskStatus struct {
	// Name of the task
	Name string `json:"name"`

	// Phase of the task (Pending, Running, Succeeded, Failed)
	Phase string `json:"phase"`

	// StartTime is when the task started execution
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the task completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Message provides additional details about the task status
	// +optional
	Message string `json:"message,omitempty"`
}

// ObservatoryStatus defines the observed state of Observatory
type ObservatoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase represents the current phase of the workflow (Pending, Running, Succeeded, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Tasks contains the status of individual tasks
	// +optional
	Tasks []TaskStatus `json:"tasks,omitempty"`

	// StartTime is when the workflow started execution
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the workflow completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// LastScheduleTime is the last time the workflow was scheduled
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Conditions represent the latest available observations of the Observatory's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=obs
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Tasks",type=integer,JSONPath=`.spec.tasks[*].name`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Observatory is the Schema for the observatories API
type Observatory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ObservatorySpec   `json:"spec,omitempty"`
	Status ObservatoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ObservatoryList contains a list of Observatory
type ObservatoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Observatory `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Observatory{}, &ObservatoryList{})
}
