package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CostCenterParameters are the configurable fields of a Dynatrace Cost Center.
type CostCenterParameters struct {
	// CostCenter is the name or key of the cost center.
	CostCenter string `json:"costCenter"`

	// Account is the optional account UUID.
	// +optional
	Account *string `json:"account,omitempty"`
}

// CostCenterObservation are the observable fields of a Dynatrace Cost Center.
type CostCenterObservation struct {
	// ID is the unique identifier of the cost center (matches costCenter key).
	ID string `json:"id,omitempty"`

	// CostCenter is the key of the cost center.
	CostCenter string `json:"costCenter,omitempty"`
}

// A CostCenterSpec defines the desired state of a Dynatrace Cost Center.
type CostCenterSpec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     CostCenterParameters `json:"forProvider"`
}

// A CostCenterStatus represents the observed state of a Dynatrace Cost Center.
type CostCenterStatus struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 CostCenterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A CostCenter is a managed resource that represents a Dynatrace Cost Center.
type CostCenter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CostCenterSpec   `json:"spec"`
	Status CostCenterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CostCenterList contains a list of CostCenter.
type CostCenterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CostCenter `json:"items"`
}
