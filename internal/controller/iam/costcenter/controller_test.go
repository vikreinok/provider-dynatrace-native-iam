package costcenter

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockDynatraceClient struct {
	dtclient.Client
	MockGetCostCenter    func(ctx context.Context, key string) (*dtclient.FieldValueDto, error)
	MockAddCostCenter    func(ctx context.Context, key string) error
	MockDeleteCostCenter func(ctx context.Context, key string) error
}

func (m *mockDynatraceClient) GetCostCenter(ctx context.Context, key string) (*dtclient.FieldValueDto, error) {
	if m.MockGetCostCenter != nil {
		return m.MockGetCostCenter(ctx, key)
	}
	return nil, nil
}

func (m *mockDynatraceClient) AddCostCenter(ctx context.Context, key string) error {
	if m.MockAddCostCenter != nil {
		return m.MockAddCostCenter(ctx, key)
	}
	return nil
}

func (m *mockDynatraceClient) DeleteCostCenter(ctx context.Context, key string) error {
	if m.MockDeleteCostCenter != nil {
		return m.MockDeleteCostCenter(ctx, key)
	}
	return nil
}

func TestCostCenterObserve(t *testing.T) {
	cases := map[string]struct {
		client   dtclient.Client
		mg       *iamv1alpha1.CostCenter
		exists   bool
		upToDate bool
	}{
		"NotFound": {
			client: &mockDynatraceClient{
				MockGetCostCenter: func(ctx context.Context, key string) (*dtclient.FieldValueDto, error) {
					return nil, &dtclient.APIError{StatusCode: 404, Message: "Not found"}
				},
			},
			mg: &iamv1alpha1.CostCenter{
				Spec: iamv1alpha1.CostCenterSpec{
					ForProvider: iamv1alpha1.CostCenterParameters{CostCenter: "CC-404"},
				},
			},
			exists:   false,
			upToDate: false,
		},
		"Found": {
			client: &mockDynatraceClient{
				MockGetCostCenter: func(ctx context.Context, key string) (*dtclient.FieldValueDto, error) {
					return &dtclient.FieldValueDto{Key: "CC-100"}, nil
				},
			},
			mg: &iamv1alpha1.CostCenter{
				Spec: iamv1alpha1.CostCenterSpec{
					ForProvider: iamv1alpha1.CostCenterParameters{CostCenter: "CC-100"},
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

func TestCostCenterCreate(t *testing.T) {
	createdKey := ""
	client := &mockDynatraceClient{
		MockAddCostCenter: func(ctx context.Context, key string) error {
			createdKey = key
			return nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.CostCenter{
		Spec: iamv1alpha1.CostCenterSpec{
			ForProvider: iamv1alpha1.CostCenterParameters{
				CostCenter: "CC-ENG-01",
			},
		},
	}

	_, err := e.Create(context.Background(), mg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if createdKey != "CC-ENG-01" {
		t.Errorf("AddCostCenter called with %s, want CC-ENG-01", createdKey)
	}
	if ext := meta.GetExternalName(mg); ext != "CC-ENG-01" {
		t.Errorf("ExternalName = %s, want CC-ENG-01", ext)
	}
}

func TestCostCenterDelete(t *testing.T) {
	deletedKey := ""
	client := &mockDynatraceClient{
		MockDeleteCostCenter: func(ctx context.Context, key string) error {
			deletedKey = key
			return nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.CostCenter{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{meta.AnnotationKeyExternalName: "CC-DEL-01"},
		},
		Spec: iamv1alpha1.CostCenterSpec{
			ForProvider: iamv1alpha1.CostCenterParameters{
				CostCenter: "CC-DEL-01",
			},
		},
	}

	_, err := e.Delete(context.Background(), mg)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if deletedKey != "CC-DEL-01" {
		t.Errorf("DeleteCostCenter called with %s, want CC-DEL-01", deletedKey)
	}
}
