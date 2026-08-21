package policyboundary

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
	MockGetBoundary    func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyBoundaryDto, error)
	MockCreateBoundary func(ctx context.Context, levelType, levelID string, boundary dtclient.PolicyBoundaryDto) (*dtclient.PolicyBoundaryDto, error)
	MockDeleteBoundary func(ctx context.Context, levelType, levelID, uuid string) error
}

func (m *mockDynatraceClient) GetBoundary(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyBoundaryDto, error) {
	if m.MockGetBoundary != nil {
		return m.MockGetBoundary(ctx, levelType, levelID, uuid)
	}
	return nil, nil
}

func (m *mockDynatraceClient) CreateBoundary(ctx context.Context, levelType, levelID string, boundary dtclient.PolicyBoundaryDto) (*dtclient.PolicyBoundaryDto, error) {
	if m.MockCreateBoundary != nil {
		return m.MockCreateBoundary(ctx, levelType, levelID, boundary)
	}
	return &dtclient.PolicyBoundaryDto{UUID: "created-bnd-uuid", Name: boundary.Name}, nil
}

func (m *mockDynatraceClient) DeleteBoundary(ctx context.Context, levelType, levelID, uuid string) error {
	if m.MockDeleteBoundary != nil {
		return m.MockDeleteBoundary(ctx, levelType, levelID, uuid)
	}
	return nil
}

func TestPolicyBoundaryObserve(t *testing.T) {
	cases := map[string]struct {
		client   dtclient.Client
		mg       *iamv1alpha1.PolicyBoundary
		exists   bool
		upToDate bool
	}{
		"NotFound": {
			client: &mockDynatraceClient{
				MockGetBoundary: func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyBoundaryDto, error) {
					return nil, &dtclient.APIError{StatusCode: 404, Message: "Not found"}
				},
			},
			mg: &iamv1alpha1.PolicyBoundary{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: "bnd-404"},
				},
				Spec: iamv1alpha1.PolicyBoundarySpec{
					ForProvider: iamv1alpha1.PolicyBoundaryParameters{Name: "NonExistent"},
				},
			},
			exists:   false,
			upToDate: false,
		},
		"Found": {
			client: &mockDynatraceClient{
				MockGetBoundary: func(ctx context.Context, levelType, levelID, uuid string) (*dtclient.PolicyBoundaryDto, error) {
					return &dtclient.PolicyBoundaryDto{
						UUID:          "bnd-100",
						Name:          "ProdBoundary",
						BoundaryQuery: "environment:management-zone = \"prod\"",
					}, nil
				},
			},
			mg: &iamv1alpha1.PolicyBoundary{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: "bnd-100"},
				},
				Spec: iamv1alpha1.PolicyBoundarySpec{
					ForProvider: iamv1alpha1.PolicyBoundaryParameters{
						Name:  "ProdBoundary",
						Query: "environment:management-zone = \"prod\"",
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

func TestPolicyBoundaryCreate(t *testing.T) {
	createdUUID := "new-bnd-uuid-555"
	client := &mockDynatraceClient{
		MockCreateBoundary: func(ctx context.Context, levelType, levelID string, boundary dtclient.PolicyBoundaryDto) (*dtclient.PolicyBoundaryDto, error) {
			return &dtclient.PolicyBoundaryDto{
				UUID: createdUUID,
				Name: boundary.Name,
			}, nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.PolicyBoundary{
		Spec: iamv1alpha1.PolicyBoundarySpec{
			ForProvider: iamv1alpha1.PolicyBoundaryParameters{
				Name:  "BoundaryTest",
				Query: "test",
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
