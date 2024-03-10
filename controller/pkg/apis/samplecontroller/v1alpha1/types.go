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

// ScaleSchedule is a specification for a ScaleSchedule resource
type ScaleSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScaleScheduleSpec   `json:"spec"`
	//Status ScaleScheduleStatus `json:"status"`
}

// ScaleScheduleSpec is the spec for a ScaleSchedule resource
type ScaleScheduleSpec struct {

    TargetRef TargetRef `json:"targetRef"`
    Schedule  []ScheduleEntry `json:"schedule"`
}

type TargetRef struct {
    DeploymentName      string `json:"deploymentName"`
    DeploymentNamespace string `json:"deploymentNamespace"`
}

type ScheduleEntry struct {
    At       string `json:"at"`
    Replicas int    `json:"replicas"`
}
// // ScaleScheduleStatus is the status for a ScaleSchedule resource
// type ScaleScheduleStatus struct {
// 	AvailableReplicas int32 `json:"availableReplicas"`
// }

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScaleScheduleList is a list of ScaleSchedule resources
type ScaleScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ScaleSchedule `json:"items"`
}
