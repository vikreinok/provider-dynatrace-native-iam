package zonev2

import (
	"context"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	managementv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/management/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockDynatraceClient struct {
	dtclient.Client
	MockGetManagementZoneV2    func(ctx context.Context, objectID string) (*dtclient.SettingsObjectItemDto, error)
	MockListManagementZonesV2  func(ctx context.Context) (*dtclient.SettingsObjectsListDto, error)
	MockCreateManagementZoneV2 func(ctx context.Context, value dtclient.ManagementZoneV2Value) (*dtclient.SettingsObjectResponseDto, error)
	MockUpdateManagementZoneV2 func(ctx context.Context, objectID string, value dtclient.ManagementZoneV2Value) error
	MockDeleteManagementZoneV2 func(ctx context.Context, objectID string) error
}

func (m *mockDynatraceClient) GetManagementZoneV2(ctx context.Context, objectID string) (*dtclient.SettingsObjectItemDto, error) {
	if m.MockGetManagementZoneV2 != nil {
		return m.MockGetManagementZoneV2(ctx, objectID)
	}
	return nil, nil
}

func (m *mockDynatraceClient) ListManagementZonesV2(ctx context.Context) (*dtclient.SettingsObjectsListDto, error) {
	if m.MockListManagementZonesV2 != nil {
		return m.MockListManagementZonesV2(ctx)
	}
	return &dtclient.SettingsObjectsListDto{}, nil
}

func (m *mockDynatraceClient) CreateManagementZoneV2(ctx context.Context, value dtclient.ManagementZoneV2Value) (*dtclient.SettingsObjectResponseDto, error) {
	if m.MockCreateManagementZoneV2 != nil {
		return m.MockCreateManagementZoneV2(ctx, value)
	}
	return &dtclient.SettingsObjectResponseDto{ObjectID: "created-obj-id"}, nil
}

func (m *mockDynatraceClient) UpdateManagementZoneV2(ctx context.Context, objectID string, value dtclient.ManagementZoneV2Value) error {
	if m.MockUpdateManagementZoneV2 != nil {
		return m.MockUpdateManagementZoneV2(ctx, objectID, value)
	}
	return nil
}

func (m *mockDynatraceClient) DeleteManagementZoneV2(ctx context.Context, objectID string) error {
	if m.MockDeleteManagementZoneV2 != nil {
		return m.MockDeleteManagementZoneV2(ctx, objectID)
	}
	return nil
}

func TestZoneV2Observe(t *testing.T) {
	type fields struct {
		client dtclient.Client
	}
	type args struct {
		mg *managementv1alpha1.ZoneV2
	}
	type want struct {
		exists   bool
		upToDate bool
		err      bool
	}

	cases := map[string]struct {
		fields fields
		args   args
		want   want
	}{
		"NotFound": {
			fields: fields{
				client: &mockDynatraceClient{
					MockGetManagementZoneV2: func(ctx context.Context, objectID string) (*dtclient.SettingsObjectItemDto, error) {
						return nil, &dtclient.APIError{StatusCode: 404, Message: "Not found"}
					},
				},
			},
			args: args{
				mg: &managementv1alpha1.ZoneV2{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: "obj-404"},
					},
					Spec: managementv1alpha1.ZoneV2Spec{
						ForProvider: managementv1alpha1.ZoneV2Parameters{Name: ptr.To("Missing")},
					},
				},
			},
			want: want{
				exists:   false,
				upToDate: false,
			},
		},
		"FoundAndUpToDate": {
			fields: fields{
				client: &mockDynatraceClient{
					MockGetManagementZoneV2: func(ctx context.Context, objectID string) (*dtclient.SettingsObjectItemDto, error) {
						return &dtclient.SettingsObjectItemDto{
							ObjectID: "obj-123",
							SchemaID: "builtin:management-zones",
							Scope:    "environment",
							Value: dtclient.ManagementZoneV2Value{
								Name:        "Test Zone",
								Description: "A test management zone",
								Rules: []dtclient.ZoneRuleDto{
									{
										Type:    "ME",
										Enabled: true,
										AttributeRule: &dtclient.AttributeRuleDto{
											EntityType: "HOST",
											AttributeConditions: []dtclient.AttributeConditionDto{
												{
													Key:         "HOST_NAME",
													Operator:    "CONTAINS",
													StringValue: ptr.To("prod"),
												},
											},
										},
									},
								},
							},
						}, nil
					},
				},
			},
			args: args{
				mg: &managementv1alpha1.ZoneV2{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: "obj-123"},
					},
					Spec: managementv1alpha1.ZoneV2Spec{
						ForProvider: managementv1alpha1.ZoneV2Parameters{
							Name:        ptr.To("Test Zone"),
							Description: ptr.To("A test management zone"),
							Rules: []managementv1alpha1.ZoneV2RulesParameters{
								{
									Rule: []managementv1alpha1.RuleParameters{
										{
											Type:    ptr.To("ME"),
											Enabled: ptr.To(true),
											AttributeRule: []managementv1alpha1.AttributeRuleParameters{
												{
													EntityType: ptr.To("HOST"),
													AttributeConditions: []managementv1alpha1.AttributeConditionsParameters{
														{
															Condition: []managementv1alpha1.AttributeConditionsConditionParameters{
																{
																	Key:         ptr.To("HOST_NAME"),
																	Operator:    ptr.To("CONTAINS"),
																	StringValue: ptr.To("prod"),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: want{
				exists:   true,
				upToDate: true,
			},
		},
		"FoundNeedsUpdate": {
			fields: fields{
				client: &mockDynatraceClient{
					MockGetManagementZoneV2: func(ctx context.Context, objectID string) (*dtclient.SettingsObjectItemDto, error) {
						return &dtclient.SettingsObjectItemDto{
							ObjectID: "obj-123",
							SchemaID: "builtin:management-zones",
							Scope:    "environment",
							Value: dtclient.ManagementZoneV2Value{
								Name:        "Old Name",
								Description: "Old Description",
							},
						}, nil
					},
				},
			},
			args: args{
				mg: &managementv1alpha1.ZoneV2{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: "obj-123"},
					},
					Spec: managementv1alpha1.ZoneV2Spec{
						ForProvider: managementv1alpha1.ZoneV2Parameters{
							Name:        ptr.To("New Name"),
							Description: ptr.To("Old Description"),
						},
					},
				},
			},
			want: want{
				exists:   true,
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if (err != nil) != tc.want.err {
				t.Errorf("Observe() error = %v, wantErr %v", err, tc.want.err)
				return
			}
			if diff := cmp.Diff(tc.want.exists, obs.ResourceExists); diff != "" {
				t.Errorf("Observe() ResourceExists mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, obs.ResourceUpToDate); diff != "" {
				t.Errorf("Observe() ResourceUpToDate mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestZoneV2Create(t *testing.T) {
	called := false
	c := &mockDynatraceClient{
		MockCreateManagementZoneV2: func(ctx context.Context, value dtclient.ManagementZoneV2Value) (*dtclient.SettingsObjectResponseDto, error) {
			called = true
			if value.Name != "Created Zone" {
				t.Errorf("expected Name 'Created Zone', got '%s'", value.Name)
			}
			return &dtclient.SettingsObjectResponseDto{ObjectID: "new-obj-id"}, nil
		},
	}

	e := &external{client: c}
	mg := &managementv1alpha1.ZoneV2{
		Spec: managementv1alpha1.ZoneV2Spec{
			ForProvider: managementv1alpha1.ZoneV2Parameters{
				Name: ptr.To("Created Zone"),
			},
		},
	}

	_, err := e.Create(context.Background(), mg)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !called {
		t.Errorf("expected CreateManagementZoneV2 to be called")
	}
	if meta.GetExternalName(mg) != "new-obj-id" {
		t.Errorf("expected external name 'new-obj-id', got '%s'", meta.GetExternalName(mg))
	}
}

func TestZoneV2Delete(t *testing.T) {
	called := false
	c := &mockDynatraceClient{
		MockDeleteManagementZoneV2: func(ctx context.Context, objectID string) error {
			called = true
			if objectID != "obj-to-delete" {
				t.Errorf("expected objectID 'obj-to-delete', got '%s'", objectID)
			}
			return nil
		},
	}

	e := &external{client: c}
	mg := &managementv1alpha1.ZoneV2{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{meta.AnnotationKeyExternalName: "obj-to-delete"},
		},
	}

	_, err := e.Delete(context.Background(), mg)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !called {
		t.Errorf("expected DeleteManagementZoneV2 to be called")
	}
}
