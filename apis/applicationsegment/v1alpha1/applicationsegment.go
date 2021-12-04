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

// CustomApplicationSegmentParameters that are not part of the ZPA API
type CustomApplicationSegmentParameters struct {
	// SegmentGroupIDRef is a reference to a SegmentGroupID so set external ID
	// +optional
	SegmentGroupIDRef *xpv1.Reference `json:"segmentGroupIDRef,omitempty"`

	// SegmentGroupIDSelector selects a reference to a SegmentGroupID so set external ID
	// +optional
	SegmentGroupIDSelector *xpv1.Selector `json:"segmentGroupIDSelector,omitempty"`
}

// A ApplicationSegmentParameters defines desired state of a ApplicationSegmentSegment
type ApplicationSegmentParameters struct {
	CustomApplicationSegmentParameters `json:",inline"`

	// bypass type
	// +kubebuilder:validation:Enum=ALWAYS;NEVER;ON_NET
	BypassType string `json:"bypassType,omitempty"`

	// config space
	// +kubebuilder:validation:Enum=DEFAULT;SIEM
	ConfigSpace string `json:"configSpace,omitempty"`

	// default idle timeout
	DefaultIdleTimeout string `json:"defaultIdleTimeout,omitempty"`

	// default max age
	DefaultMaxAge string `json:"defaultMaxAge,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// domain names
	DomainNames []string `json:"domainNames"`

	// double encrypt
	DoubleEncrypt *bool `json:"doubleEncrypt,omitempty"`

	// enabled
	Enabled *bool `json:"enabled,omitempty"`

	// health check type
	// +kubebuilder:validation:Enum=DEFAULT;NONE
	HealthCheckType string `json:"healthCheckType,omitempty"`

	// health reporting
	// +kubebuilder:validation:Enum=NONE;ON_ACCESS;CONTINUOUS
	HealthReporting string `json:"healthReporting,omitempty"`

	// icmp access type
	// +kubebuilder:validation:Enum=PING_TRACEROUTING;PING;NONE
	IcmpAccessType string `json:"icmpAccessType,omitempty"`

	// ip anchored
	IPAnchored *bool `json:"ipAnchored,omitempty"`

	// is cname enabled
	IsCnameEnabled *bool `json:"isCnameEnabled,omitempty"`

	// passive health enabled
	PassiveHealthEnabled *bool `json:"passiveHealthEnabled,omitempty"`

	// segment group Id
	SegmentGroupID *string `json:"segmentGroupID,omitempty"`

	// tcp port ranges
	TCPPortRanges []string `json:"tcpPortRanges,omitempty"`

	// udp port ranges
	UDPPortRanges []string `json:"udpPortRanges,omitempty"`

	// CustomerID The unique identifier of the ZPA tenant.
	// +kubebuilder:validation:Required
	CustomerID string `json:"customerID"`
}

// A ApplicationSegmentSpec defines the desired state of a ApplicationSegment.
type ApplicationSegmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationSegmentParameters `json:"forProvider"`
}

// A ApplicationSegmentStatus represents the status of a ApplicationSegment.
type ApplicationSegmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          Observation `json:"atProvider,omitempty"`
}

// Observation are the observable fields of a ApplicationSegment.
type Observation struct {
	CreationTime string `json:"creationTime,omitempty"`
	ModifiedBy   string `json:"modifiedBy,omitempty"`
	ModifiedTime string `json:"modifiedTime,omitempty"`
	ID           string `json:"id,omitempty"`
}

// +kubebuilder:object:root=true

// A ApplicationSegment is the schema for ZPA ApplicationSegments API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type ApplicationSegment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSegmentSpec   `json:"spec"`
	Status ApplicationSegmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationSegmentList contains a list of ApplicationSegment
type ApplicationSegmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationSegment `json:"items"`
}
