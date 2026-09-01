package zonev2

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	managementv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/management/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
	"github.com/vikreinok/provider-dynatrace-native-iam/internal/controller/helper"
)

const (
	errNotZoneV2    = "managed resource is not a ZoneV2 custom resource"
	errGetZoneV2    = "cannot get ZoneV2 from Dynatrace API"
	errCreateZoneV2 = "cannot create ZoneV2 in Dynatrace API"
	errUpdateZoneV2 = "cannot update ZoneV2 in Dynatrace API"
	errDeleteZoneV2 = "cannot delete ZoneV2 in Dynatrace API"
)

// SetupGated adds a controller that reconciles ZoneV2 managed resources with safe-start support.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupGatedManagedController(
		mgr,
		o,
		managementv1alpha1.ZoneV2GroupVersionKind,
		managementv1alpha1.ZoneV2GroupKind,
		&managementv1alpha1.ZoneV2{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

// Setup adds a controller that reconciles ZoneV2 managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return helper.SetupManagedController(
		mgr,
		o,
		managementv1alpha1.ZoneV2GroupVersionKind,
		managementv1alpha1.ZoneV2GroupKind,
		&managementv1alpha1.ZoneV2{},
		&helper.DynatraceConnector{
			Kube: mgr.GetClient(),
			NewExternalClientFn: func(client dtclient.Client) managed.ExternalClient {
				return &external{client: client}
			},
		},
	)
}

type external struct {
	client dtclient.Client
}


func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*managementv1alpha1.ZoneV2)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotZoneV2)
	}

	objectID := meta.GetExternalName(cr)
	var item *dtclient.SettingsObjectItemDto
	var err error

	if objectID != "" && objectID != cr.GetName() {
		item, err = e.client.GetManagementZoneV2(ctx, objectID)
	} else if cr.Spec.ForProvider.Name != nil {
		// Attempt to match by name
		list, listErr := e.client.ListManagementZonesV2(ctx)
		if listErr != nil {
			return managed.ExternalObservation{}, errors.Wrap(listErr, errGetZoneV2)
		}
		for _, it := range list.Items {
			if it.Value.Name == *cr.Spec.ForProvider.Name {
				matched := it
				item = &matched
				break
			}
		}
	}

	if err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetZoneV2)
	}

	if item == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	meta.SetExternalName(cr, item.ObjectID)
	cr.Status.AtProvider = toZoneV2Observation(item)

	cr.SetConditions(xpv2.Available())

	upToDate := isZoneV2UpToDate(cr.Spec.ForProvider, &item.Value)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*managementv1alpha1.ZoneV2)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotZoneV2)
	}

	cr.SetConditions(xpv2.Creating())

	val := toManagementZoneV2Value(cr.Spec.ForProvider)

	created, err := e.client.CreateManagementZoneV2(ctx, val)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateZoneV2)
	}

	meta.SetExternalName(cr, created.ObjectID)
	cr.Status.AtProvider.ID = &created.ObjectID

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*managementv1alpha1.ZoneV2)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotZoneV2)
	}

	objectID := meta.GetExternalName(cr)
	if objectID == "" {
		return managed.ExternalUpdate{}, errors.New("cannot update ZoneV2 without external-name")
	}

	val := toManagementZoneV2Value(cr.Spec.ForProvider)

	if err := e.client.UpdateManagementZoneV2(ctx, objectID, val); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateZoneV2)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*managementv1alpha1.ZoneV2)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotZoneV2)
	}

	cr.SetConditions(xpv2.Deleting())

	objectID := meta.GetExternalName(cr)
	if objectID == "" {
		return managed.ExternalDelete{}, nil
	}

	if err := e.client.DeleteManagementZoneV2(ctx, objectID); err != nil {
		if dtclient.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteZoneV2)
	}

	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(ctx context.Context) error {
	return nil
}

func toManagementZoneV2Value(p managementv1alpha1.ZoneV2Parameters) dtclient.ManagementZoneV2Value {
	name := ""
	if p.Name != nil {
		name = *p.Name
	}
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	legacyID := ""
	if p.LegacyID != nil {
		legacyID = *p.LegacyID
	}

	rules := make([]dtclient.ZoneRuleDto, 0)
	for _, rp := range p.Rules {
		for _, r := range rp.Rule {
			dto := dtclient.ZoneRuleDto{
				Enabled: true,
			}
			if r.Type != nil {
				dto.Type = *r.Type
			}
			if r.Enabled != nil {
				dto.Enabled = *r.Enabled
			}
			if r.EntitySelector != nil {
				dto.EntitySelector = *r.EntitySelector
			}

			if len(r.AttributeRule) > 0 {
				ar := r.AttributeRule[0]
				arDto := &dtclient.AttributeRuleDto{
					HostToPgpropagation:                        ar.HostToPgpropagation,
					PgToHostPropagation:                        ar.PgToHostPropagation,
					PgToServicePropagation:                     ar.PgToServicePropagation,
					ServiceToHostPropagation:                   ar.ServiceToHostPropagation,
					ServiceToPgpropagation:                     ar.ServiceToPgpropagation,
					AzureToPgpropagation:                       ar.AzureToPgpropagation,
					AzureToServicePropagation:                  ar.AzureToServicePropagation,
					CustomDeviceGroupToCustomDevicePropagation: ar.CustomDeviceGroupToCustomDevicePropagation,
				}
				if ar.EntityType != nil {
					arDto.EntityType = *ar.EntityType
				}
				for _, ac := range ar.AttributeConditions {
					for _, c := range ac.Condition {
						cDto := dtclient.AttributeConditionDto{
							StringValue:      c.StringValue,
							EnumValue:        c.EnumValue,
							IntegerValue:     c.IntegerValue,
							EntityID:         c.EntityID,
							DynamicKey:       c.DynamicKey,
							DynamicKeySource: c.DynamicKeySource,
							CaseSensitive:    c.CaseSensitive,
							Tag:              c.Tag,
						}
						if c.Key != nil {
							cDto.Key = *c.Key
						}
						if c.Operator != nil {
							cDto.Operator = *c.Operator
						}
						arDto.AttributeConditions = append(arDto.AttributeConditions, cDto)
					}
				}
				dto.AttributeRule = arDto
			}

			if len(r.DimensionRule) > 0 {
				dr := r.DimensionRule[0]
				drDto := &dtclient.DimensionRuleDto{}
				if dr.AppliesTo != nil {
					drDto.AppliesTo = *dr.AppliesTo
				}
				for _, dc := range dr.DimensionConditions {
					for _, c := range dc.Condition {
						cDto := dtclient.DimensionConditionDto{
							Key: c.Key,
						}
						if c.ConditionType != nil {
							cDto.ConditionType = *c.ConditionType
						}
						if c.RuleMatcher != nil {
							cDto.RuleMatcher = *c.RuleMatcher
						}
						if c.Value != nil {
							cDto.Value = *c.Value
						}
						drDto.DimensionConditions = append(drDto.DimensionConditions, cDto)
					}
				}
				dto.DimensionRule = drDto
			}

			rules = append(rules, dto)
		}
	}

	return dtclient.ManagementZoneV2Value{
		Name:        name,
		Description: desc,
		LegacyID:    legacyID,
		Rules:       rules,
	}
}

func toZoneV2Observation(obj *dtclient.SettingsObjectItemDto) managementv1alpha1.ZoneV2Observation {
	obs := managementv1alpha1.ZoneV2Observation{
		ID:          &obj.ObjectID,
		Name:        &obj.Value.Name,
		Description: &obj.Value.Description,
		LegacyID:    &obj.Value.LegacyID,
	}

	rulesList := make([]managementv1alpha1.RuleObservation, 0, len(obj.Value.Rules))
	for _, r := range obj.Value.Rules {
		rObs := managementv1alpha1.RuleObservation{
			Type:           &r.Type,
			Enabled:        &r.Enabled,
			EntitySelector: &r.EntitySelector,
		}

		if r.AttributeRule != nil {
			ar := r.AttributeRule
			arObs := managementv1alpha1.AttributeRuleObservation{
				EntityType:                                 &ar.EntityType,
				HostToPgpropagation:                        ar.HostToPgpropagation,
				PgToHostPropagation:                        ar.PgToHostPropagation,
				PgToServicePropagation:                     ar.PgToServicePropagation,
				ServiceToHostPropagation:                   ar.ServiceToHostPropagation,
				ServiceToPgpropagation:                     ar.ServiceToPgpropagation,
				AzureToPgpropagation:                       ar.AzureToPgpropagation,
				AzureToServicePropagation:                  ar.AzureToServicePropagation,
				CustomDeviceGroupToCustomDevicePropagation: ar.CustomDeviceGroupToCustomDevicePropagation,
			}
			condObsList := make([]managementv1alpha1.AttributeConditionsConditionObservation, 0, len(ar.AttributeConditions))
			for _, c := range ar.AttributeConditions {
				condObsList = append(condObsList, managementv1alpha1.AttributeConditionsConditionObservation{
					Key:              &c.Key,
					Operator:         &c.Operator,
					StringValue:      c.StringValue,
					EnumValue:        c.EnumValue,
					IntegerValue:     c.IntegerValue,
					EntityID:         c.EntityID,
					DynamicKey:       c.DynamicKey,
					DynamicKeySource: c.DynamicKeySource,
					CaseSensitive:    c.CaseSensitive,
					Tag:              c.Tag,
				})
			}
			if len(condObsList) > 0 {
				arObs.AttributeConditions = []managementv1alpha1.AttributeConditionsObservation{
					{Condition: condObsList},
				}
			}
			rObs.AttributeRule = []managementv1alpha1.AttributeRuleObservation{arObs}
		}

		if r.DimensionRule != nil {
			dr := r.DimensionRule
			drObs := managementv1alpha1.DimensionRuleObservation{
				AppliesTo: &dr.AppliesTo,
			}
			condObsList := make([]managementv1alpha1.DimensionConditionsConditionObservation, 0, len(dr.DimensionConditions))
			for _, c := range dr.DimensionConditions {
				condObsList = append(condObsList, managementv1alpha1.DimensionConditionsConditionObservation{
					ConditionType: &c.ConditionType,
					Key:           c.Key,
					RuleMatcher:   &c.RuleMatcher,
					Value:         &c.Value,
				})
			}
			if len(condObsList) > 0 {
				drObs.DimensionConditions = []managementv1alpha1.DimensionConditionsObservation{
					{Condition: condObsList},
				}
			}
			rObs.DimensionRule = []managementv1alpha1.DimensionRuleObservation{drObs}
		}

		rulesList = append(rulesList, rObs)
	}

	if len(rulesList) > 0 {
		obs.Rules = []managementv1alpha1.ZoneV2RulesObservation{
			{Rule: rulesList},
		}
	}

	return obs
}

func isZoneV2UpToDate(p managementv1alpha1.ZoneV2Parameters, val *dtclient.ManagementZoneV2Value) bool {
	if p.Name != nil && *p.Name != val.Name {
		return false
	}
	if p.Description != nil && *p.Description != val.Description {
		return false
	}
	if p.LegacyID != nil && *p.LegacyID != val.LegacyID {
		return false
	}

	desired := toManagementZoneV2Value(p)
	// Compare rules
	if len(desired.Rules) != len(val.Rules) {
		return false
	}

	desiredJSON, err1 := json.Marshal(desired.Rules)
	actualJSON, err2 := json.Marshal(val.Rules)
	if err1 != nil || err2 != nil {
		return reflect.DeepEqual(desired.Rules, val.Rules)
	}
	return string(desiredJSON) == string(actualJSON)
}
