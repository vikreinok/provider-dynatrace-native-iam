// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

type HostEntityInitParameters struct {
	// The type of the entity, e.g. "HOST"
	Type *string `json:"type,omitempty"`

	// The name of the entity, e.g. "ip-172-31-21-198.eu-central-1.compute.internal"
	Name *string `json:"name,omitempty"`
}

type HostEntityObservation struct {
	// The resolved Dynatrace ID of the entity, e.g. "HOST-D6E60CF7996C2E61"
	EntityID *string `json:"entityId,omitempty"`

	// The tags of the entity in Dynatrace.
	Tags []HostEntityTag `json:"tags,omitempty"`
}

type HostEntityTag struct {
	Context              *string `json:"context,omitempty"`
	Key                  *string `json:"key,omitempty"`
	Value                *string `json:"value,omitempty"`
	StringRepresentation *string `json:"stringRepresentation,omitempty"`
}

type HostEntityParameters struct {
	// The type of the entity, e.g. "HOST"
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`

	// The name of the entity, e.g. "ip-172-31-21-198.eu-central-1.compute.internal"
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`
}

// HostEntitySpec defines the desired state of HostEntity
type HostEntitySpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     HostEntityParameters     `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	InitProvider                    HostEntityInitParameters `json:"initProvider,omitempty"`
}

// HostEntityStatus defines the observed state of HostEntity.
type HostEntityStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 HostEntityObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// HostEntity is the Schema for the HostEntities API.
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
type HostEntity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.type) || (has(self.initProvider) && has(self.initProvider.type))",message="spec.forProvider.type is a required parameter"
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.name) || (has(self.initProvider) && has(self.initProvider.name))",message="spec.forProvider.name is a required parameter"
	Spec   HostEntitySpec   `json:"spec"`
	Status HostEntityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HostEntityList contains a list of HostEntities
type HostEntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostEntity `json:"items"`
}
