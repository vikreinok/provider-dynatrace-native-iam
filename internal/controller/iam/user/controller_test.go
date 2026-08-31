package user

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockDynatraceClient struct {
	dtclient.Client
	MockGetUser    func(ctx context.Context, email string) (*dtclient.UserDto, error)
	MockCreateUser func(ctx context.Context, email string, groups []string) error
	MockDeleteUser func(ctx context.Context, email string) error
}

func (m *mockDynatraceClient) GetUser(ctx context.Context, email string) (*dtclient.UserDto, error) {
	if m.MockGetUser != nil {
		return m.MockGetUser(ctx, email)
	}
	return nil, nil
}

func (m *mockDynatraceClient) CreateUser(ctx context.Context, email string, groups []string) error {
	if m.MockCreateUser != nil {
		return m.MockCreateUser(ctx, email, groups)
	}
	return nil
}

func (m *mockDynatraceClient) DeleteUser(ctx context.Context, email string) error {
	if m.MockDeleteUser != nil {
		return m.MockDeleteUser(ctx, email)
	}
	return nil
}

func TestUserResolveGroups(t *testing.T) {
	groupUUID := "b35f4a41-5a4d-45ea-b7a6-514ab5660fb7"
	scheme := runtime.NewScheme()
	_ = iamv1alpha1.SchemeBuilder.AddToScheme(scheme)

	kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&iamv1alpha1.Group{
			ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
			Status: iamv1alpha1.GroupStatus{
				AtProvider: iamv1alpha1.GroupObservation{ID: groupUUID},
			},
		},
		&iamv1alpha1.Group{
			ObjectMeta: metav1.ObjectMeta{Name: "not-ready-group"},
			Status: iamv1alpha1.GroupStatus{
				AtProvider: iamv1alpha1.GroupObservation{ID: ""},
			},
		},
	).Build()

	cases := map[string]struct {
		kubeClient client.Client
		groups     []string
		want       []string
		wantErr    bool
	}{
		"LiteralUUID": {
			kubeClient: kubeClient,
			groups:     []string{groupUUID},
			want:       []string{groupUUID},
		},
		"GroupRefName_Ready": {
			kubeClient: kubeClient,
			groups:     []string{"my-group"},
			want:       []string{groupUUID},
		},
		"GroupRefName_NotReady": {
			kubeClient: kubeClient,
			groups:     []string{"not-ready-group"},
			wantErr:    true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kubeClient}
			got, err := e.resolveGroups(context.Background(), tc.groups)
			if (err != nil) != tc.wantErr {
				t.Fatalf("resolveGroups() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if len(got) != len(tc.want) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("got[%d] = %s, want %s", i, got[i], tc.want[i])
				}
			}
		})
	}
}
