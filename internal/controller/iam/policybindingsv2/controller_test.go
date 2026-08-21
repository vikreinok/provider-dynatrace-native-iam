package policybindingsv2

import (
	"context"
	"testing"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockDynatraceClient struct {
	dtclient.Client
	MockGetPolicyBindingsForGroup func(ctx context.Context, levelType, levelID, groupUUID string) (*dtclient.PolicyBindingsDto, error)
	MockSetPolicyBinding          func(ctx context.Context, levelType, levelID, policyUUID, groupUUID string, binding dtclient.AppendLevelPolicyBindingForGroupDto) error
	MockDeletePolicyBinding       func(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) error
}

func (m *mockDynatraceClient) GetPolicyBindingsForGroup(ctx context.Context, levelType, levelID, groupUUID string) (*dtclient.PolicyBindingsDto, error) {
	if m.MockGetPolicyBindingsForGroup != nil {
		return m.MockGetPolicyBindingsForGroup(ctx, levelType, levelID, groupUUID)
	}
	return nil, nil
}

func (m *mockDynatraceClient) SetPolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string, binding dtclient.AppendLevelPolicyBindingForGroupDto) error {
	if m.MockSetPolicyBinding != nil {
		return m.MockSetPolicyBinding(ctx, levelType, levelID, policyUUID, groupUUID, binding)
	}
	return nil
}

func (m *mockDynatraceClient) DeletePolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) error {
	if m.MockDeletePolicyBinding != nil {
		return m.MockDeletePolicyBinding(ctx, levelType, levelID, policyUUID, groupUUID)
	}
	return nil
}

func TestPolicyBindingsV2Observe(t *testing.T) {
	groupUUID := "grp-binding-1"
	cases := map[string]struct {
		client   dtclient.Client
		mg       *iamv1alpha1.PolicyBindingsV2
		exists   bool
		upToDate bool
	}{
		"NotFound": {
			client: &mockDynatraceClient{
				MockGetPolicyBindingsForGroup: func(ctx context.Context, levelType, levelID, groupUUID string) (*dtclient.PolicyBindingsDto, error) {
					return &dtclient.PolicyBindingsDto{
						PolicyBindings: []dtclient.BindingItem{},
					}, nil
				},
			},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						Group:  &groupUUID,
						Policy: []iamv1alpha1.PolicyBindingItem{{ID: "pol-1"}},
					},
				},
			},
			exists:   false,
			upToDate: false,
		},
		"FoundAndUpToDate": {
			client: &mockDynatraceClient{
				MockGetPolicyBindingsForGroup: func(ctx context.Context, levelType, levelID, groupUUID string) (*dtclient.PolicyBindingsDto, error) {
					return &dtclient.PolicyBindingsDto{
						PolicyBindings: []dtclient.BindingItem{
							{ID: "pol-1", Boundaries: []string{"bnd-1"}},
						},
					}, nil
				},
			},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						Group: &groupUUID,
						Policy: []iamv1alpha1.PolicyBindingItem{
							{ID: "pol-1", Boundaries: []string{"bnd-1"}},
						},
					},
				},
			},
			exists:   true,
			upToDate: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			obs, err := e.Observe(context.Background(), tc.mg)
			if err != nil {
				t.Fatalf("Observe() error = %v", err)
			}
			if obs.ResourceExists != tc.exists {
				t.Errorf("ResourceExists = %v, want %v", obs.ResourceExists, tc.exists)
			}
			if obs.ResourceUpToDate != tc.upToDate {
				t.Errorf("ResourceUpToDate = %v, want %v", obs.ResourceUpToDate, tc.upToDate)
			}
		})
	}
}
