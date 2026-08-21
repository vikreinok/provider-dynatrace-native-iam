package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyBoundaryParameters are the configurable fields of a Dynatrace Policy Boundary.
type PolicyBoundaryParameters struct {
	// Name is the display name of the policy boundary.
	Name string `json:"name"`

	// Query is the boundary query restricting scope (e.g. environment:management-zone startsWith "SV-...").
	Query string `json:"query"`

	// Account is the account UUID if scoped to an account.
	// +optional
	Account *string `json:"account,omitempty"`

	// Environment is the environment ID if scoped to a specific environment.
	// +optional
	Environment *string `json:"environment,omitempty"`
}

// PolicyBoundaryObservation are the observable fields of a Dynatrace Policy Boundary.
type PolicyBoundaryObservation struct {
	// ID is the unique UUID of the boundary.
	ID string `json:"id,omitempty"`

	// UUID is the unique UUID of the boundary.
	UUID string `json:"uuid,omitempty"`

	// LevelType is the scope level type (account or environment).
	LevelType string `json:"levelType,omitempty"`

	// LevelID is the level target ID (account UUID or environment ID).
	LevelID string `json:"levelId,omitempty"`
}

// A PolicyBoundarySpec defines the desired state of a Dynatrace Policy Boundary.
type PolicyBoundarySpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     PolicyBoundaryParameters `json:"forProvider"`
}

// A PolicyBoundaryStatus represents the observed state of a Dynatrace Policy Boundary.
type PolicyBoundaryStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 PolicyBoundaryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A PolicyBoundary is a managed resource that represents a Dynatrace IAM Policy Boundary.
type PolicyBoundary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyBoundarySpec   `json:"spec"`
	Status PolicyBoundaryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyBoundaryList contains a list of PolicyBoundary.
type PolicyBoundaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyBoundary `json:"items"`
}
