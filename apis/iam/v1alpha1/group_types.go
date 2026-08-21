package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupPermission defines a permission granted to the group.
type GroupPermission struct {
	// PermissionName is the name of the permission (e.g. account-company-info, tenant-viewer).
	PermissionName string `json:"permissionName"`

	// Scope is the scope target (account UUID, tenant ID, or management-zone ID).
	Scope string `json:"scope"`

	// ScopeType is the type of scope (account, tenant, management-zone).
	ScopeType string `json:"scopeType"`
}

// GroupParameters are the configurable fields of a Dynatrace IAM Group.
type GroupParameters struct {
	// Name is the display name of the user group.
	Name string `json:"name"`

	// Description is an optional description of the user group.
	// +optional
	Description *string `json:"description,omitempty"`

	// FederatedAttributeValues is a list of SAML group attribute values that map to this group.
	// +optional
	FederatedAttributeValues []string `json:"federatedAttributeValues,omitempty"`

	// Permissions is an optional list of permissions assigned to this group.
	// +optional
	Permissions []GroupPermission `json:"permissions,omitempty"`
}

// GroupObservation are the observable fields of a Dynatrace IAM Group.
type GroupObservation struct {
	// ID is the UUID of the group in Dynatrace.
	ID string `json:"id,omitempty"`

	// Owner indicates if the group is LOCAL or SAML federated.
	Owner string `json:"owner,omitempty"`

	// Hidden indicates whether the group is hidden in the Dynatrace UI.
	Hidden bool `json:"hidden,omitempty"`

	// CreatedAt is the ISO timestamp when the group was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// UpdatedAt is the ISO timestamp when the group was last modified.
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// A GroupSpec defines the desired state of a Dynatrace IAM Group.
type GroupSpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     GroupParameters `json:"forProvider"`
}

// A GroupStatus represents the observed state of a Dynatrace IAM Group.
type GroupStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 GroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A Group is a managed resource that represents a Dynatrace IAM Group.
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec"`
	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of Group.
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}
