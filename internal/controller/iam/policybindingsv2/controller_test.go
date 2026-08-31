package policybindingsv2

import (
	"context"
	"testing"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

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

func TestIsUUID(t *testing.T) {
	cases := map[string]struct {
		input string
		want  bool
	}{
		"ValidUUID": {
			input: "b35f4a41-5a4d-45ea-b7a6-514ab5660fb7",
			want:  true,
		},
		"UppercaseUUID": {
			input: "B35F4A41-5A4D-45EA-B7A6-514AB5660FB7",
			want:  true,
		},
		"InvalidLength": {
			input: "b35f4a41-5a4d-45ea-b7a6",
			want:  false,
		},
		"InvalidChars": {
			input: "g35f4a41-5a4d-45ea-b7a6-514ab5660fb7",
			want:  false,
		},
		"ResourceName": {
			input: "sv-service-purpose-role",
			want:  false,
		},
		"Empty": {
			input: "",
			want:  false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := isUUID(tc.input); got != tc.want {
				t.Errorf("isUUID(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestPolicyBindingsV2Observe(t *testing.T) {
	groupUUID := "b35f4a41-5a4d-45ea-b7a6-514ab5660fb7"
	policyUUID := "89e4c1a2-3f12-4211-9a10-123456789abc"
	boundaryUUID := "f1e2d3c4-b5a6-7890-1234-56789abcdef0"

	scheme := runtime.NewScheme()
	_ = iamv1alpha1.SchemeBuilder.AddToScheme(scheme)

	cases := map[string]struct {
		kubeClient client.Client
		dtClient   dtclient.Client
		mg         *iamv1alpha1.PolicyBindingsV2
		exists     bool
		upToDate   bool
		wantErr    bool
	}{
		"LiteralUUID_NotFound": {
			dtClient: &mockDynatraceClient{
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
						Policy: []iamv1alpha1.PolicyBindingItem{{ID: policyUUID}},
					},
				},
			},
			exists:   false,
			upToDate: false,
		},
		"LiteralUUID_FoundAndUpToDate": {
			dtClient: &mockDynatraceClient{
				MockGetPolicyBindingsForGroup: func(ctx context.Context, levelType, levelID, groupUUID string) (*dtclient.PolicyBindingsDto, error) {
					return &dtclient.PolicyBindingsDto{
						PolicyBindings: []dtclient.BindingItem{
							{ID: policyUUID, Boundaries: []string{boundaryUUID}},
						},
					}, nil
				},
			},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						Group: &groupUUID,
						Policy: []iamv1alpha1.PolicyBindingItem{
							{ID: policyUUID, Boundaries: []string{boundaryUUID}},
						},
					},
				},
			},
			exists:   true,
			upToDate: true,
		},
		"GroupRef_ResolvedReady": {
			kubeClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&iamv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "my-group-cr"},
					Status: iamv1alpha1.GroupStatus{
						AtProvider: iamv1alpha1.GroupObservation{ID: groupUUID},
					},
				},
				&iamv1alpha1.Policy{
					ObjectMeta: metav1.ObjectMeta{Name: "my-policy-cr"},
					Status: iamv1alpha1.PolicyStatus{
						AtProvider: iamv1alpha1.PolicyObservation{ID: policyUUID},
					},
				},
			).Build(),
			dtClient: &mockDynatraceClient{
				MockGetPolicyBindingsForGroup: func(ctx context.Context, levelType, levelID, gUUID string) (*dtclient.PolicyBindingsDto, error) {
					if gUUID != groupUUID {
						t.Errorf("GetPolicyBindingsForGroup groupUUID = %s, want %s", gUUID, groupUUID)
					}
					return &dtclient.PolicyBindingsDto{
						PolicyBindings: []dtclient.BindingItem{
							{ID: policyUUID},
						},
					}, nil
				},
			},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						GroupRef: &xpv2.Reference{Name: "my-group-cr"},
						Policy:   []iamv1alpha1.PolicyBindingItem{{ID: "my-policy-cr"}},
					},
				},
			},
			exists:   true,
			upToDate: true,
		},
		"GroupRef_WaitingForReady": {
			kubeClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&iamv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "my-group-cr"},
					Status: iamv1alpha1.GroupStatus{
						AtProvider: iamv1alpha1.GroupObservation{ID: ""},
					},
				},
			).Build(),
			dtClient: &mockDynatraceClient{},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						GroupRef: &xpv2.Reference{Name: "my-group-cr"},
						Policy:   []iamv1alpha1.PolicyBindingItem{{ID: policyUUID}},
					},
				},
			},
			wantErr: true,
		},
		"Group_NonUUID_WaitingForReady": {
			kubeClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&iamv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{Name: "sv-group-name"},
					Status: iamv1alpha1.GroupStatus{
						AtProvider: iamv1alpha1.GroupObservation{ID: ""},
					},
				},
			).Build(),
			dtClient: &mockDynatraceClient{},
			mg: &iamv1alpha1.PolicyBindingsV2{
				Spec: iamv1alpha1.PolicyBindingsV2Spec{
					ForProvider: iamv1alpha1.PolicyBindingsV2Parameters{
						Group:  ptr("sv-group-name"),
						Policy: []iamv1alpha1.PolicyBindingItem{{ID: policyUUID}},
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kubeClient, client: tc.dtClient}
			obs, err := e.Observe(context.Background(), tc.mg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Observe() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
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

func ptr[T any](v T) *T {
	return &v
}
