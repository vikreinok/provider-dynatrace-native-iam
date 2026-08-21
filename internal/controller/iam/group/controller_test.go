package group

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
	MockGetGroup    func(ctx context.Context, uuid string) (*dtclient.GroupDto, error)
	MockListGroups  func(ctx context.Context) (*dtclient.GroupListDto, error)
	MockCreateGroup func(ctx context.Context, group dtclient.GroupDto) (*dtclient.GroupDto, error)
	MockUpdateGroup func(ctx context.Context, uuid string, group dtclient.GroupDto) error
	MockDeleteGroup func(ctx context.Context, uuid string) error
}

func (m *mockDynatraceClient) GetGroup(ctx context.Context, uuid string) (*dtclient.GroupDto, error) {
	if m.MockGetGroup != nil {
		return m.MockGetGroup(ctx, uuid)
	}
	return nil, nil
}

func (m *mockDynatraceClient) ListGroups(ctx context.Context) (*dtclient.GroupListDto, error) {
	if m.MockListGroups != nil {
		return m.MockListGroups(ctx)
	}
	return &dtclient.GroupListDto{}, nil
}

func (m *mockDynatraceClient) CreateGroup(ctx context.Context, group dtclient.GroupDto) (*dtclient.GroupDto, error) {
	if m.MockCreateGroup != nil {
		return m.MockCreateGroup(ctx, group)
	}
	return &dtclient.GroupDto{UUID: "created-uuid", Name: group.Name}, nil
}

func (m *mockDynatraceClient) UpdateGroup(ctx context.Context, uuid string, group dtclient.GroupDto) error {
	if m.MockUpdateGroup != nil {
		return m.MockUpdateGroup(ctx, uuid, group)
	}
	return nil
}

func (m *mockDynatraceClient) DeleteGroup(ctx context.Context, uuid string) error {
	if m.MockDeleteGroup != nil {
		return m.MockDeleteGroup(ctx, uuid)
	}
	return nil
}

func TestGroupObserve(t *testing.T) {
	type fields struct {
		client dtclient.Client
	}
	type args struct {
		mg *iamv1alpha1.Group
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
					MockGetGroup: func(ctx context.Context, uuid string) (*dtclient.GroupDto, error) {
						return nil, &dtclient.APIError{StatusCode: 404, Message: "Not found"}
					},
				},
			},
			args: args{
				mg: &iamv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: "grp-404"},
					},
					Spec: iamv1alpha1.GroupSpec{
						ForProvider: iamv1alpha1.GroupParameters{Name: "Missing"},
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
					MockGetGroup: func(ctx context.Context, uuid string) (*dtclient.GroupDto, error) {
						return &dtclient.GroupDto{
							UUID:        "grp-123",
							Name:        "Engineers",
							Description: "Dev Team",
						}, nil
					},
				},
			},
			args: args{
				mg: &iamv1alpha1.Group{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{meta.AnnotationKeyExternalName: "grp-123"},
					},
					Spec: iamv1alpha1.GroupSpec{
						ForProvider: iamv1alpha1.GroupParameters{
							Name:        "Engineers",
							Description: func() *string { s := "Dev Team"; return &s }(),
						},
					},
				},
			},
			want: want{
				exists:   true,
				upToDate: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
			obs, err := e.Observe(context.Background(), tc.args.mg)
			if (err != nil) != tc.want.err {
				t.Fatalf("Observe() error = %v, wantErr %v", err, tc.want.err)
			}
			if obs.ResourceExists != tc.want.exists {
				t.Errorf("ResourceExists = %v, want %v", obs.ResourceExists, tc.want.exists)
			}
			if obs.ResourceUpToDate != tc.want.upToDate {
				t.Errorf("ResourceUpToDate = %v, want %v", obs.ResourceUpToDate, tc.want.upToDate)
			}
		})
	}
}

func TestGroupCreate(t *testing.T) {
	createdUUID := "new-group-uuid-999"
	client := &mockDynatraceClient{
		MockCreateGroup: func(ctx context.Context, group dtclient.GroupDto) (*dtclient.GroupDto, error) {
			return &dtclient.GroupDto{
				UUID:  createdUUID,
				Name:  group.Name,
				Owner: "LOCAL",
			}, nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.Group{
		Spec: iamv1alpha1.GroupSpec{
			ForProvider: iamv1alpha1.GroupParameters{
				Name: "NewGroup",
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

func TestGroupDelete(t *testing.T) {
	deletedUUID := ""
	client := &mockDynatraceClient{
		MockDeleteGroup: func(ctx context.Context, uuid string) error {
			deletedUUID = uuid
			return nil
		},
	}

	e := &external{client: client}
	mg := &iamv1alpha1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{meta.AnnotationKeyExternalName: "del-uuid-1"},
		},
	}

	_, err := e.Delete(context.Background(), mg)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if deletedUUID != "del-uuid-1" {
		t.Errorf("DeleteGroup called with %s, want del-uuid-1", deletedUUID)
	}
}
