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

// A ServerGroupParameters defines desired state of a ServerSegment
type ServerGroupParameters struct {

	// enabled
	Enabled *bool `json:"enabled,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// ip anchored
	IPAnchored *bool `json:"ipAnchored,omitempty"`

	// config space
	// +kubebuilder:validation:Enum=DEFAULT;SIEM
	ConfigSpace string `json:"configSpace,omitempty"`

	// Defaults to false.
	DynamicDiscovery bool `json:"dynamicDiscovery"`

	// app connector groups
	// +required
	AppConnectorGroups []string `json:"appConnectorGroups"`

	// Name for ServerGroup.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// A ServerGroupSpec defines the desired state of a ServerGroup.
type ServerGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ServerGroupParameters `json:"forProvider"`
}

// A ServerGroupStatus represents the status of a ServerGroup.
type ServerGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          Observation `json:"atProvider,omitempty"`
}

// Observation are the observable fields of a ServerGroup.
type Observation struct {
	CreationTime string `json:"creationTime,omitempty"`
	ModifiedBy   string `json:"modifiedBy,omitempty"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
	ID           string `json:"id,omitempty"`
}

// +kubebuilder:object:root=true

// A ServerGroup is the schema for ZPA ServerGroups API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type ServerGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerGroupSpec   `json:"spec"`
	Status ServerGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerGroupList contains a list of ServerGroup
type ServerGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerGroup `json:"items"`
}
