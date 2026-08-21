package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyBindingItem represents a single policy attachment within PolicyBindingsV2.
type PolicyBindingItem struct {
	// ID is the unique UUID of the policy being bound.
	ID string `json:"id"`

	// Boundaries is a list of policy boundary UUIDs to restrict the policy scope.
	// +optional
	Boundaries []string `json:"boundaries,omitempty"`

	// Parameters are key-value policy parameters (e.g. service-purpose).
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// Metadata are optional key-value metadata tags for the binding.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PolicyBindingsV2Parameters are the configurable fields of Dynatrace PolicyBindingsV2.
type PolicyBindingsV2Parameters struct {
	// Group is the UUID of the IAM group to bind policies to.
	// +optional
	Group *string `json:"group,omitempty"`

	// GroupRef references a Group managed resource to retrieve its UUID.
	// +optional
	GroupRef *xpv2.Reference `json:"groupRef,omitempty"`

	// GroupSelector selects a Group managed resource to retrieve its UUID.
	// +optional
	GroupSelector *xpv2.Selector `json:"groupSelector,omitempty"`

	// Environment is the environment ID if bound at environment level (e.g. owt43371).
	// +optional
	Environment *string `json:"environment,omitempty"`

	// Account is the account UUID if bound at account level.
	// +optional
	Account *string `json:"account,omitempty"`

	// Policy is the list of policies, boundaries, and parameters to bind to the group.
	// +optional
	Policy []PolicyBindingItem `json:"policy,omitempty"`
}

// PolicyBindingsV2Observation are the observable fields of Dynatrace PolicyBindingsV2.
type PolicyBindingsV2Observation struct {
	// ID is the synthetic unique identifier of this binding set.
	ID string `json:"id,omitempty"`

	// Group is the resolved UUID of the group.
	Group string `json:"group,omitempty"`

	// Environment is the environment ID where policies are bound.
	Environment string `json:"environment,omitempty"`

	// Account is the account UUID where policies are bound.
	Account string `json:"account,omitempty"`
}

// A PolicyBindingsV2Spec defines the desired state of Dynatrace PolicyBindingsV2.
type PolicyBindingsV2Spec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     PolicyBindingsV2Parameters `json:"forProvider"`
}

// A PolicyBindingsV2Status represents the observed state of Dynatrace PolicyBindingsV2.
type PolicyBindingsV2Status struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 PolicyBindingsV2Observation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A PolicyBindingsV2 is a managed resource that represents Dynatrace IAM Policy Bindings V2.
type PolicyBindingsV2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyBindingsV2Spec   `json:"spec"`
	Status PolicyBindingsV2Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyBindingsV2List contains a list of PolicyBindingsV2.
type PolicyBindingsV2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyBindingsV2 `json:"items"`
}
