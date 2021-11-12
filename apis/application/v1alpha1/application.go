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

// CustomApplicationParameters that are not part of the ZPA API
type CustomApplicationParameters struct {
	// SegmentGroupIDRef is a reference to a SegmentGroupID so set external ID
	// +optional
	SegmentGroupIDRef *xpv1.Reference `json:"segmentGroupIDRef,omitempty"`

	// SegmentGroupIDSelector selects a reference to a SegmentGroupID so set external ID
	// +optional
	SegmentGroupIDSelector *xpv1.Selector `json:"segmentGroupIDSelector,omitempty"`
}

// A ApplicationParameters defines desired state of a Application
type ApplicationParameters struct {
	CustomApplicationParameters `json:",inline"`

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
	TCPPortRanges []string `json:"tcpPortRanges"`

	// udp port ranges
	UDPPortRanges []string `json:"udpPortRanges"`

	// CustomerID The unique identifier of the ZPA tenant.
	// +kubebuilder:validation:Required
	CustomerID int64 `json:"customerID"`
}

// A ApplicationSpec defines the desired state of a Application.
type ApplicationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationParameters `json:"forProvider"`
}

// A ApplicationStatus represents the status of a Application.
type ApplicationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A Application is the schema for ZPA Applications API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,zpa}
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}
