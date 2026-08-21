package policy

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockDynatraceClient struct {
	dtclient.Client
	MockGetPolicy    func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyDto, error)
	MockListPolicies func(ctx context.Context, levelType, levelID string) (*dtclient.PolicyListDto, error)
	MockCreatePolicy func(ctx context.Context, levelType, levelID string, policy dtclient.PolicyDto) (*dtclient.PolicyDto, error)
	MockUpdatePolicy func(ctx context.Context, levelType, levelID, uuid string, policy dtclient.PolicyDto) (*dtclient.PolicyDto, error)
	MockDeletePolicy func(ctx context.Context, levelType, levelID, uuid string) error
}

func (m *mockDynatraceClient) GetPolicy(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyDto, error) {
	if m.MockGetPolicy != nil {
		return m.MockGetPolicy(ctx, levelType, levelID, uuid)
	}
	return nil, nil
}

func (m *mockDynatraceClient) ListPolicies(ctx context.Context, levelType, levelID string) (*dtclient.PolicyListDto, error) {
	if m.MockListPolicies != nil {
		return m.MockListPolicies(ctx, levelType, levelID)
	}
	return &dtclient.PolicyListDto{}, nil
}

func (m *mockDynatraceClient) CreatePolicy(ctx context.Context, levelType, levelID string, policy dtclient.PolicyDto) (*dtclient.PolicyDto, error) {
	if m.MockCreatePolicy != nil {
		return m.MockCreatePolicy(ctx, levelType, levelID, policy)
	}
	return &dtclient.PolicyDto{UUID: "created-pol-uuid", Name: policy.Name}, nil
}

func (m *mockDynatraceClient) UpdatePolicy(ctx context.Context, levelType, levelID, uuid string, policy dtclient.PolicyDto) (*dtclient.PolicyDto, error) {
	if m.MockUpdatePolicy != nil {
		return m.MockUpdatePolicy(ctx, levelType, levelID, uuid, policy)
	}
	return &dtclient.PolicyDto{UUID: uuid, Name: policy.Name}, nil
}

func (m *mockDynatraceClient) DeletePolicy(ctx context.Context, levelType, levelID, uuid string) error {
	if m.MockDeletePolicy != nil {
		return m.MockDeletePolicy(ctx, levelType, levelID, uuid)
	}
	return nil
}

func TestPolicyObserve(t *testing.T) {
	cases := map[string]struct {
		client   dtclient.Client
		mg       *iamv1alpha1.Policy
		exists   bool
		upToDate bool
	}{
		"NotFound": {
			client: &mockDynatraceClient{
				MockGetPolicy: func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyDto, error) {
					return nil, &dtclient.APIError{StatusCode: 404, Message: "Not found"}
				},
			},
			mg: &iamv1alpha1.Policy{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: "pol-404"},
				},
				Spec: iamv1alpha1.PolicySpec{
					ForProvider: iamv1alpha1.PolicyParameters{Name: "NonExistent"},
				},
			},
			exists:   false,
			upToDate: false,
		},
		"FoundAndUpToDate": {
			client: &mockDynatraceClient{
				MockGetPolicy: func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyDto, error) {
					return &dtclient.PolicyDto{
						UUID:           "pol-123",
						Name:           "ReadOnly",
						StatementQuery: "ALLOW settings:objects:read;",
					}, nil
				},
			},
			mg: &iamv1alpha1.Policy{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: "pol-123"},
				},
				Spec: iamv1alpha1.PolicySpec{
					ForProvider: iamv1alpha1.PolicyParameters{
						Name:           "ReadOnly",
						StatementQuery: "ALLOW settings:objects:read;",
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

func TestPolicyCreate(t *testing.T) {
	createdUUID := "new-pol-uuid-100"
	client := &mockDynatraceClient{
		MockCreatePolicy: func(ctx context.Context, levelType, levelID string, policy dtclient.PolicyDto) (*dtclient.PolicyDto, error) {
			return &dtclient.PolicyDto{
				UUID: createdUUID,
				Name: policy.Name,
			}, nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.Policy{
		Spec: iamv1alpha1.PolicySpec{
			ForProvider: iamv1alpha1.PolicyParameters{
				Name:           "NewPolicy",
				StatementQuery: "ALLOW *;",
			},
		},
	}

	_, err := e.Create(context.Background(), mg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if ext := meta.GetExternalName(mg); ext != createdUUID {
		t.Errorf("ExternalName = %s, want %s", ext, createdUUID)
	}
	if diff := cmp.Diff(createdUUID, mg.Status.AtProvider.ID); diff != "" {
		t.Errorf("Status.AtProvider.ID mismatch (-want +got):\n%s", diff)
	}
}
