package v1alpha1

import (
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AttributeConditionsConditionParameters represents condition within attribute rule.
type AttributeConditionsConditionParameters struct {
	// CaseSensitive indicates if comparison is case sensitive.
	// +optional
	CaseSensitive *bool `json:"caseSensitive,omitempty"`

	// DynamicKey specifies dynamic key.
	// +optional
	DynamicKey *string `json:"dynamicKey,omitempty"`

	// DynamicKeySource specifies dynamic key source.
	// +optional
	DynamicKeySource *string `json:"dynamicKeySource,omitempty"`

	// EntityID specifies value.
	// +optional
	EntityID *string `json:"entityId,omitempty"`

	// EnumValue specifies enum value.
	// +optional
	EnumValue *string `json:"enumValue,omitempty"`

	// IntegerValue specifies integer value.
	// +optional
	IntegerValue *float64 `json:"integerValue,omitempty"`

	// Key specifies condition key.
	Key *string `json:"key,omitempty"`

	// Operator specifies comparison operator (e.g. BEGINS_WITH, CONTAINS, ENDS_WITH, EQUALS, EXISTS, etc).
	Operator *string `json:"operator,omitempty"`

	// StringValue specifies string value.
	// +optional
	StringValue *string `json:"stringValue,omitempty"`

	// Tag specifies format: [CONTEXT]tagKey:tagValue.
	// +optional
	Tag *string `json:"tag,omitempty"`
}

// AttributeConditionsConditionObservation represents condition observation within attribute rule.
type AttributeConditionsConditionObservation struct {
	CaseSensitive    *bool    `json:"caseSensitive,omitempty"`
	DynamicKey       *string  `json:"dynamicKey,omitempty"`
	DynamicKeySource *string  `json:"dynamicKeySource,omitempty"`
	EntityID         *string  `json:"entityId,omitempty"`
	EnumValue        *string  `json:"enumValue,omitempty"`
	IntegerValue     *float64 `json:"integerValue,omitempty"`
	Key              *string  `json:"key,omitempty"`
	Operator         *string  `json:"operator,omitempty"`
	StringValue      *string  `json:"stringValue,omitempty"`
	Tag              *string  `json:"tag,omitempty"`
}

// AttributeConditionsParameters holds conditions list.
type AttributeConditionsParameters struct {
	Condition []AttributeConditionsConditionParameters `json:"condition,omitempty"`
}

// AttributeConditionsObservation holds conditions list observation.
type AttributeConditionsObservation struct {
	Condition []AttributeConditionsConditionObservation `json:"condition,omitempty"`
}

// AttributeRuleParameters defines entity filtering rules.
type AttributeRuleParameters struct {
	// AttributeConditions holds list of attribute conditions.
	// +optional
	AttributeConditions []AttributeConditionsParameters `json:"attributeConditions,omitempty"`

	// AzureToPgpropagation applies to process groups connected to matching Azure entities.
	// +optional
	AzureToPgpropagation *bool `json:"azureToPgpropagation,omitempty"`

	// AzureToServicePropagation applies to services provided by matching Azure entities.
	// +optional
	AzureToServicePropagation *bool `json:"azureToServicePropagation,omitempty"`

	// CustomDeviceGroupToCustomDevicePropagation applies to custom devices in a custom device group.
	// +optional
	CustomDeviceGroupToCustomDevicePropagation *bool `json:"customDeviceGroupToCustomDevicePropagation,omitempty"`

	// EntityType specifies monitored entity type (e.g. HOST, SERVICE, PROCESS_GROUP, etc).
	EntityType *string `json:"entityType,omitempty"`

	// HostToPgpropagation applies to processes running on matching hosts.
	// +optional
	HostToPgpropagation *bool `json:"hostToPgpropagation,omitempty"`

	// PgToHostPropagation applies to underlying hosts of matching process groups.
	// +optional
	PgToHostPropagation *bool `json:"pgToHostPropagation,omitempty"`

	// PgToServicePropagation applies to all services provided by the process groups.
	// +optional
	PgToServicePropagation *bool `json:"pgToServicePropagation,omitempty"`

	// ServiceToHostPropagation applies to underlying hosts of matching services.
	// +optional
	ServiceToHostPropagation *bool `json:"serviceToHostPropagation,omitempty"`

	// ServiceToPgpropagation applies to underlying process groups of matching services.
	// +optional
	ServiceToPgpropagation *bool `json:"serviceToPgpropagation,omitempty"`
}

// AttributeRuleObservation defines observed entity filtering rules.
type AttributeRuleObservation struct {
	AttributeConditions                        []AttributeConditionsObservation `json:"attributeConditions,omitempty"`
	AzureToPgpropagation                       *bool                            `json:"azureToPgpropagation,omitempty"`
	AzureToServicePropagation                  *bool                            `json:"azureToServicePropagation,omitempty"`
	CustomDeviceGroupToCustomDevicePropagation *bool                            `json:"customDeviceGroupToCustomDevicePropagation,omitempty"`
	EntityType                                 *string                          `json:"entityType,omitempty"`
	HostToPgpropagation                        *bool                            `json:"hostToPgpropagation,omitempty"`
	PgToHostPropagation                        *bool                            `json:"pgToHostPropagation,omitempty"`
	PgToServicePropagation                     *bool                            `json:"pgToServicePropagation,omitempty"`
	ServiceToHostPropagation                   *bool                            `json:"serviceToHostPropagation,omitempty"`
	ServiceToPgpropagation                     *bool                            `json:"serviceToPgpropagation,omitempty"`
}

// DimensionConditionsConditionParameters defines a condition for dimensional rules.
type DimensionConditionsConditionParameters struct {
	// ConditionType: DIMENSION, LOG_FILE_NAME, METRIC_KEY.
	ConditionType *string `json:"conditionType,omitempty"`

	// Key is the dimension key.
	// +optional
	Key *string `json:"key,omitempty"`

	// RuleMatcher: BEGINS_WITH, EQUALS.
	RuleMatcher *string `json:"ruleMatcher,omitempty"`

	// Value is the dimension value.
	Value *string `json:"value,omitempty"`
}

// DimensionConditionsConditionObservation defines observed dimension condition.
type DimensionConditionsConditionObservation struct {
	ConditionType *string `json:"conditionType,omitempty"`
	Key           *string `json:"key,omitempty"`
	RuleMatcher   *string `json:"ruleMatcher,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// DimensionConditionsParameters holds dimension conditions.
type DimensionConditionsParameters struct {
	Condition []DimensionConditionsConditionParameters `json:"condition,omitempty"`
}

// DimensionConditionsObservation holds observed dimension conditions.
type DimensionConditionsObservation struct {
	Condition []DimensionConditionsConditionObservation `json:"condition,omitempty"`
}

// DimensionRuleParameters defines dimensional data filtering rules.
type DimensionRuleParameters struct {
	// AppliesTo: ANY, LOG, METRIC.
	AppliesTo *string `json:"appliesTo,omitempty"`

	// DimensionConditions holds conditions.
	// +optional
	DimensionConditions []DimensionConditionsParameters `json:"dimensionConditions,omitempty"`
}

// DimensionRuleObservation defines observed dimensional data filtering rules.
type DimensionRuleObservation struct {
	AppliesTo           *string                          `json:"appliesTo,omitempty"`
	DimensionConditions []DimensionConditionsObservation `json:"dimensionConditions,omitempty"`
}

// RuleParameters defines a management zone rule.
type RuleParameters struct {
	// AttributeRule defines entity rule.
	// +optional
	AttributeRule []AttributeRuleParameters `json:"attributeRule,omitempty"`

	// DimensionRule defines dimensional rule.
	// +optional
	DimensionRule []DimensionRuleParameters `json:"dimensionRule,omitempty"`

	// Enabled indicates whether the rule is active.
	Enabled *bool `json:"enabled,omitempty"`

	// EntitySelector defines advanced entity selector.
	// +optional
	EntitySelector *string `json:"entitySelector,omitempty"`

	// Type: ME, DIMENSION, SELECTOR.
	Type *string `json:"type,omitempty"`
}

// RuleObservation defines observed management zone rule.
type RuleObservation struct {
	AttributeRule  []AttributeRuleObservation `json:"attributeRule,omitempty"`
	DimensionRule  []DimensionRuleObservation `json:"dimensionRule,omitempty"`
	Enabled        *bool                      `json:"enabled,omitempty"`
	EntitySelector *string                    `json:"entitySelector,omitempty"`
	Type           *string                    `json:"type,omitempty"`
}

// ZoneV2RulesParameters holds rules for the management zone.
type ZoneV2RulesParameters struct {
	// Rule is a list of management zone rules.
	// +optional
	Rule []RuleParameters `json:"rule,omitempty"`
}

// ZoneV2RulesObservation holds observed rules for the management zone.
type ZoneV2RulesObservation struct {
	Rule []RuleObservation `json:"rule,omitempty"`
}

// ZoneV2Parameters are the configurable fields of a Dynatrace Management Zone V2.
type ZoneV2Parameters struct {
	// Name of the management zone.
	Name *string `json:"name,omitempty"`

	// Description of the management zone.
	// +optional
	Description *string `json:"description,omitempty"`

	// LegacyID is the ID when referred to by Config REST API V1.
	// +optional
	LegacyID *string `json:"legacyId,omitempty"`

	// Rules for the management zone.
	// +optional
	Rules []ZoneV2RulesParameters `json:"rules,omitempty"`
}

// ZoneV2Observation are the observable fields of a Dynatrace Management Zone V2.
type ZoneV2Observation struct {
	// ID is the Settings 2.0 object ID.
	ID *string `json:"id,omitempty"`

	// Name of the management zone.
	Name *string `json:"name,omitempty"`

	// Description of the management zone.
	Description *string `json:"description,omitempty"`

	// LegacyID is the ID when referred to by Config REST API V1.
	LegacyID *string `json:"legacyId,omitempty"`

	// Rules for the management zone.
	Rules []ZoneV2RulesObservation `json:"rules,omitempty"`
}

// A ZoneV2Spec defines the desired state of a Dynatrace Management Zone V2.
type ZoneV2Spec struct {
	xpv2.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     ZoneV2Parameters `json:"forProvider"`
}

// A ZoneV2Status represents the observed state of a Dynatrace Management Zone V2.
type ZoneV2Status struct {
	xpv2.ManagedResourceStatus `json:",inline"`
	AtProvider                 ZoneV2Observation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dynatrace}
// +kubebuilder:storageversion

// A ZoneV2 is a managed resource that represents a Dynatrace Management Zone V2.
type ZoneV2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZoneV2Spec   `json:"spec"`
	Status ZoneV2Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZoneV2List contains a list of ZoneV2.
type ZoneV2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZoneV2 `json:"items"`
}
