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

// A SegmentParameters defines desired state of a Segment
type SegmentParameters struct {
	CustomSegmentParameters `json:",inline"`

	// config space
	// +kubebuilder:validation:Enum=DEFAULT;SIEM
	ConfigSpace string `json:"configSpace,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// enabled
	Enabled *bool `json:"enabled,omitempty"`

	// policy migrated
	PolicyMigrated *bool `json:"policyMigrated,omitempty"`

	// tcp keep alive enabled
	TCPKeepAliveEnabled string `json:"tcpKeepAliveEnabled,omitempty"`

	// CustomerID The unique identifier of the ZPA tenant.
	// +kubebuilder:validation:Required
	CustomerID int64 `json:"customerID"`
}

// A SegmentSpec defines the desired state of a Segment.
type SegmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SegmentParameters `json:"forProvider"`
}

// A SegmentStatus represents the status of a Segment.
type SegmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A Segment is the schema for ZPA Segments API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type Segment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SegmentSpec   `json:"spec"`
	Status SegmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SegmentList contains a list of Segment
type SegmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Segment `json:"items"`
}
