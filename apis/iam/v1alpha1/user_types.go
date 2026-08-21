package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserParameters are the configurable fields of a Dynatrace IAM User.
type UserParameters struct {
	// Email is the email address of the user.
	Email string `json:"email"`

	// Groups is the optional list of group UUIDs to assign this user to.
	// +optional
	Groups []string `json:"groups,omitempty"`
}

// UserObservation are the observable fields of a Dynatrace IAM User.
type UserObservation struct {
	// UID is the unique user ID in Dynatrace.
	UID string `json:"uid,omitempty"`

	// Email is the email address of the user.
	Email string `json:"email,omitempty"`

	// Name is the first name of the user.
	Name string `json:"name,omitempty"`

	// Surname is the last name of the user.
	Surname string `json:"surname,omitempty"`

	// UserType is the type of user (e.g. LOCAL, SAML).
	UserType string `json:"userType,omitempty"`

	// Groups is the list of assigned group UUIDs.
	// +optional
	Groups []string `json:"groups,omitempty"`
}

// A UserSpec defines the desired state of a Dynatrace IAM User.
type UserSpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     UserParameters `json:"forProvider"`
}

// A UserStatus represents the observed state of a Dynatrace IAM User.
type UserStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A User is a managed resource that represents a Dynatrace IAM User.
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of User.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}
