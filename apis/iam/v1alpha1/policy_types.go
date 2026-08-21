package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyParameters are the configurable fields of a Dynatrace IAM Policy.
type PolicyParameters struct {
	// Name is the display name of the policy.
	Name string `json:"name"`

	// Description is an optional description of the policy.
	// +optional
	Description *string `json:"description,omitempty"`

	// StatementQuery is the DQL policy statement (e.g. ALLOW settings:objects:read;).
	StatementQuery string `json:"statementQuery"`

	// Account is the account UUID if scoped to an account.
	// +optional
	Account *string `json:"account,omitempty"`

	// Environment is the environment ID if scoped to a specific environment.
	// +optional
	Environment *string `json:"environment,omitempty"`

	// Tags is an optional list of tags associated with the policy.
	// +optional
	Tags []string `json:"tags,omitempty"`
}

// PolicyObservation are the observable fields of a Dynatrace IAM Policy.
type PolicyObservation struct {
	// ID is the unique UUID of the policy.
	ID string `json:"id,omitempty"`

	// UUID is the unique UUID of the policy.
	UUID string `json:"uuid,omitempty"`

	// LevelType is the scope level type (global, account, or environment).
	LevelType string `json:"levelType,omitempty"`

	// LevelID is the level target ID (global, account UUID, or environment ID).
	LevelID string `json:"levelId,omitempty"`
}

// A PolicySpec defines the desired state of a Dynatrace IAM Policy.
type PolicySpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     PolicyParameters `json:"forProvider"`
}

// A PolicyStatus represents the observed state of a Dynatrace IAM Policy.
type PolicyStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 PolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A Policy is a managed resource that represents a Dynatrace IAM Policy.
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy.
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
