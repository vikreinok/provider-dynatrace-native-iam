package entity

import (
	"context"
	"errors"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"

	hostv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/host/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockClient struct {
	dtclient.Client
	lookupFn func(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error)
}

func (m *mockClient) LookupHostEntity(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error) {
	if m.lookupFn != nil {
		return m.lookupFn(ctx, entityType, entityName)
	}
	return "", nil, 0, nil
}

func TestObserve(t *testing.T) {
	tagContext := "ENVIRONMENT"
	tagKey := "primary_tags.company.com/service"
	tagVal := "SV-XYZ2"
	tagStr := "[Environment]primary_tags.company.com/service:SV-XYZ2"

	cases := map[string]struct {
		client   dtclient.Client
		cr       *hostv1alpha1.HostEntity
		wantID   *string
		wantTag  int
		wantCond xpv2.ConditionType
		wantErr  bool
	}{
		"DeletedResource": {
			client: &mockClient{},
			cr: &hostv1alpha1.HostEntity{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
			},
			wantErr: false,
		},
		"MissingName": {
			client: &mockClient{},
			cr: &hostv1alpha1.HostEntity{
				Spec: hostv1alpha1.HostEntitySpec{
					ForProvider: hostv1alpha1.HostEntityParameters{
						Type: ptr.To("HOST"),
					},
				},
			},
			wantCond: xpv2.TypeSynced,
			wantErr:  false,
		},
		"FoundSingleEntity": {
			client: &mockClient{
				lookupFn: func(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error) {
					return "HOST-12345", []dtclient.HostEntityTagDto{
						{
							Context:              &tagContext,
							Key:                  &tagKey,
							Value:                &tagVal,
							StringRepresentation: &tagStr,
						},
					}, 1, nil
				},
			},
			cr: &hostv1alpha1.HostEntity{
				Spec: hostv1alpha1.HostEntitySpec{
					ForProvider: hostv1alpha1.HostEntityParameters{
						Name: ptr.To("CityX"),
						Type: ptr.To("HOST"),
					},
				},
			},
			wantID:   ptr.To("HOST-12345"),
			wantTag:  1,
			wantCond: xpv2.TypeReady,
			wantErr:  false,
		},
		"FoundMultipleEntities": {
			client: &mockClient{
				lookupFn: func(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error) {
					return "HOST-12345", nil, 3, nil
				},
			},
			cr: &hostv1alpha1.HostEntity{
				Spec: hostv1alpha1.HostEntitySpec{
					ForProvider: hostv1alpha1.HostEntityParameters{
						Name: ptr.To("City*"),
						Type: ptr.To("HOST"),
					},
				},
			},
			wantID:   ptr.To("HOST-12345"),
			wantTag:  0,
			wantCond: xpv2.TypeReady,
			wantErr:  false,
		},
		"NotFoundEntity": {
			client: &mockClient{
				lookupFn: func(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error) {
					return "", nil, 0, nil
				},
			},
			cr: &hostv1alpha1.HostEntity{
				Spec: hostv1alpha1.HostEntitySpec{
					ForProvider: hostv1alpha1.HostEntityParameters{
						Name: ptr.To("non-existent-host"),
					},
				},
			},
			wantID:   nil,
			wantCond: xpv2.TypeReady,
			wantErr:  false,
		},
		"LookupError": {
			client: &mockClient{
				lookupFn: func(ctx context.Context, entityType, entityName string) (string, []dtclient.HostEntityTagDto, int, error) {
					return "", nil, 0, errors.New("network failure")
				},
			},
			cr: &hostv1alpha1.HostEntity{
				Spec: hostv1alpha1.HostEntitySpec{
					ForProvider: hostv1alpha1.HostEntityParameters{
						Name: ptr.To("CityX"),
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.client}
			obs, err := e.Observe(context.Background(), tc.cr)

			if (err != nil) != tc.wantErr {
				t.Fatalf("Observe() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if meta.WasDeleted(tc.cr) {
				if obs.ResourceExists {
					t.Errorf("expected ResourceExists = false for deleted resource")
				}
				return
			}

			if tc.wantID != nil {
				if tc.cr.Status.AtProvider.EntityID == nil || *tc.cr.Status.AtProvider.EntityID != *tc.wantID {
					t.Errorf("expected entity ID %v, got %v", tc.wantID, tc.cr.Status.AtProvider.EntityID)
				}
				if len(tc.cr.Status.AtProvider.Tags) != tc.wantTag {
					t.Errorf("expected %d tags, got %d", tc.wantTag, len(tc.cr.Status.AtProvider.Tags))
				}
				if meta.GetExternalName(tc.cr) != *tc.wantID {
					t.Errorf("expected external name %s, got %s", *tc.wantID, meta.GetExternalName(tc.cr))
				}
			} else if tc.cr.Spec.ForProvider.Name != nil && *tc.cr.Spec.ForProvider.Name == "non-existent-host" {
				if tc.cr.Status.AtProvider.EntityID != nil {
					t.Errorf("expected nil entity ID, got %v", tc.cr.Status.AtProvider.EntityID)
				}
			}
		})
	}
}

func TestLifecycleNoops(t *testing.T) {
	e := &external{client: &mockClient{}}
	cr := &hostv1alpha1.HostEntity{}

	if _, err := e.Create(context.Background(), cr); err != nil {
		t.Errorf("Create unexpected error: %v", err)
	}

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Errorf("Update unexpected error: %v", err)
	}

	if _, err := e.Delete(context.Background(), cr); err != nil {
		t.Errorf("Delete unexpected error: %v", err)
	}

	if err := e.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect unexpected error: %v", err)
	}
}
