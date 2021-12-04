/*
Copyright 2021 The Crossplane Authors.
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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomSegmentParameters that are not part of the ZPA API
type CustomSegmentParameters struct{}

// SegmentGroupParameters defines desired state of a Segment
type SegmentGroupParameters struct {
	CustomSegmentParameters `json:",inline"`

	// config space
	// +kubebuilder:validation:Enum=DEFAULT;SIEM
	ConfigSpace string `json:"configSpace,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// enabled
	// +kubebuilder:validation:Enum=true
	Enabled *bool `json:"enabled"`

	// policy migrated
	PolicyMigrated *bool `json:"policyMigrated,omitempty"`

	// tcp keep alive enabled
	// +kubebuilder:validation:Enum="0";"1"
	TCPKeepAliveEnabled string `json:"tcpKeepAliveEnabled,omitempty"`

	// CustomerID The unique identifier of the ZPA tenant.
	// +kubebuilder:validation:Required
	CustomerID string `json:"customerID"`
}

// A SegmentGroupSpec defines the desired state of a SegmentGroup.
type SegmentGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SegmentGroupParameters `json:"forProvider"`
}

// A SegmentGroupStatus represents the status of a SegmentGroup.
type SegmentGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          Observation `json:"atProvider,omitempty"`
}

// Observation are the observable fields of a SegmentGroup.
type Observation struct {
	CreationTime   string `json:"creationTime,omitempty"`
	ModifiedBy     string `json:"modifiedBy,omitempty"`
	ModifiedTime   string `json:"modifiedTime,omitempty"`
	ID             string `json:"id,omitempty"`
	PolicyMigrated bool   `json:"policyMigrated,omitempty"`
}

// +kubebuilder:object:root=true

// A SegmentGroup is the schema for ZPA SegmentGroups API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type SegmentGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SegmentGroupSpec   `json:"spec"`
	Status SegmentGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SegmentGroupList contains a list of SegmentGroup
type SegmentGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SegmentGroup `json:"items"`
}
