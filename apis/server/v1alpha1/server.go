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

// CustomServerParameters that are not part of the ZPA API
type CustomServerParameters struct {
	// ServerGroupRefs is a reference to a ServerGroup so set external ID
	// +optional
	ServerGroupRefs []xpv1.Reference `json:"serverGroupRefs,omitempty"`

	// ServerGroupSelector selects a reference to a ServerGroup so set external ID
	// +optional
	ServerGroupSelector *xpv1.Selector `json:"serverGroupSelector,omitempty"`
}

// A ServerParameters defines desired state of a ServerSegment
type ServerParameters struct {
	CustomServerParameters `json:",inline"`

	// config space
	// +kubebuilder:validation:Enum=DEFAULT;SIEM
	ConfigSpace string `json:"configSpace,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// dynamic discovery
	DynamicDiscovery *bool `json:"dynamicDiscovery,omitempty"`

	// enabled
	Enabled *bool `json:"enabled,omitempty"`

	// Domain or IP-Address
	Address string `json:"address,omitempty"`

	// +kubebuilder:validation:Required
	Name *string `json:"name"`

	// server group ids
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-zpa/apis/servergroup/v1alpha1.ServerGroup
	// +crossplane:generate:reference:refFieldName=ServerGroupRefs
	// +crossplane:generate:reference:selectorFieldName=ServerGroupSelector
	ServerGroups []string `json:"serverGroups,omitempty"`
}

// A ServerSpec defines the desired state of a Server.
type ServerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ServerParameters `json:"forProvider"`
}

// A ServerStatus represents the status of a Server.
type ServerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          Observation `json:"atProvider,omitempty"`
}

// Observation are the observable fields of a Server.
type Observation struct {
	CreationTime string `json:"creationTime,omitempty"`
	ModifiedBy   string `json:"modifiedBy,omitempty"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
	ID           string `json:"id,omitempty"`
}

// +kubebuilder:object:root=true

// A Server is the schema for ZPA Servers API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec"`
	Status ServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}
