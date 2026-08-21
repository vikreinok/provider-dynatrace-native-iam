package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceUserParameters are the configurable fields of a Dynatrace Service User.
type ServiceUserParameters struct {
	// Name is the display name of the service user.
	Name string `json:"name"`

	// Description is an optional description of the service user.
	// +optional
	Description *string `json:"description,omitempty"`

	// Groups is the optional list of group UUIDs to assign to this service user.
	// +optional
	Groups []string `json:"groups,omitempty"`
}

// ServiceUserObservation are the observable fields of a Dynatrace Service User.
type ServiceUserObservation struct {
	// UID is the unique service user UUID in Dynatrace.
	UID string `json:"uid,omitempty"`

	// Email is the generated email address of the service user.
	Email string `json:"email,omitempty"`
}

// A ServiceUserSpec defines the desired state of a Dynatrace Service User.
type ServiceUserSpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     ServiceUserParameters `json:"forProvider"`
}

// A ServiceUserStatus represents the observed state of a Dynatrace Service User.
type ServiceUserStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 ServiceUserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A ServiceUser is a managed resource that represents a Dynatrace Service User.
type ServiceUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceUserSpec   `json:"spec"`
	Status ServiceUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceUserList contains a list of ServiceUser.
type ServiceUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceUser `json:"items"`
}
