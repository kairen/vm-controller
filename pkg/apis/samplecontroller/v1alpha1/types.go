/*
Copyright 2017 The Kubernetes Authors.

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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is a specification for a Foo resource
type Foo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooSpec   `json:"spec"`
	Status FooStatus `json:"status"`
}

// FooSpec is the spec for a Foo resource
type FooSpec struct {
	DeploymentName string `json:"deploymentName"`
	Replicas       *int32 `json:"replicas"`
}

// FooStatus is the status for a Foo resource
type FooStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type FooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Foo `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VM is a specification for a VM resource
type VM struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VMSpec   `json:"spec"`
	Status VMStatus `json:"status"`
}

// VMSpec is the spec for a VM resource
type VMSpec struct {
	VMName   string `json:"vmName"`
	CPU      int32  `json:"cpu"`
	Memory   int32  `json:"memory"`
	DiskSize int32  `json:"diskSize"`
}

type VMPhase string

const (
	VMNone        VMPhase = ""
	VMPending     VMPhase = "Pending"
	VMActive      VMPhase = "Active"
	VMFailed      VMPhase = "Failed"
	VMTerminating VMPhase = "Terminating"
)

// VMStatus is the status for a VM resource
type VMStatus struct {
	Phase          VMPhase     `json:"phase"`
	Reason         string      `json:"reason,omitempty"`
	ID             string      `json:"vmId,omitempty"`
	CPUUtilization int32       `json:"cpuUtilization,omitempty"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VMList is a list of Foo resources
type VMList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VM `json:"items"`
}
