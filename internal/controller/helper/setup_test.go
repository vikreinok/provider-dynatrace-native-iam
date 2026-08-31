package helper

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/gate"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/vikreinok/provider-dynatrace-native-iam/apis"
	iamv1alpha1 "github.com/vikreinok/provider-dynatrace-native-iam/apis/iam/v1alpha1"
	"github.com/vikreinok/provider-dynatrace-native-iam/apis/v1alpha1"
	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

type mockExternalClient struct {
	managed.ExternalClient
}

func TestDynatraceConnector_Connect(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = apis.AddToScheme(scheme)

	credsData, _ := json.Marshal(dtclient.RawCredentials{
		AccountID:    "acc-123",
		ClientID:     "dt0s01.sample",
		ClientSecret: "sample-secret",
	})

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dt-creds",
			Namespace: "crossplane-system",
		},
		Data: map[string][]byte{
			"credentials": credsData,
		},
	}

	cpc := &v1alpha1.ClusterProviderConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
		Spec: v1alpha1.ProviderConfigSpec{
			Credentials: v1alpha1.ProviderCredentials{
				Source: xpv2.CredentialsSourceSecret,
				CommonCredentialSelectors: xpv2.CommonCredentialSelectors{
					SecretRef: &xpv2.SecretKeySelector{
						SecretReference: xpv2.SecretReference{
							Name:      "dt-creds",
							Namespace: "crossplane-system",
						},
						Key: "credentials",
					},
				},
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret, cpc).Build()
		called := false
		conn := &DynatraceConnector{
			Kube: client,
			NewExternalClientFn: func(c dtclient.Client) managed.ExternalClient {
				called = true
				return &mockExternalClient{}
			},
		}

		policy := &iamv1alpha1.Policy{
			ObjectMeta: metav1.ObjectMeta{Name: "test-policy"},
		}

		ext, err := conn.Connect(context.Background(), policy)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ext == nil {
			t.Fatalf("expected non-nil ExternalClient")
		}
		if !called {
			t.Fatalf("expected NewExternalClientFn to be called")
		}
	})

	t.Run("ProviderConfigNotFound", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		conn := &DynatraceConnector{
			Kube: client,
			NewExternalClientFn: func(c dtclient.Client) managed.ExternalClient {
				return &mockExternalClient{}
			},
		}

		policy := &iamv1alpha1.Policy{
			ObjectMeta: metav1.ObjectMeta{Name: "test-policy"},
		}

		_, err := conn.Connect(context.Background(), policy)
		if err == nil {
			t.Fatalf("expected error when ClusterProviderConfig is missing, got nil")
		}
	})
}

func TestSetupGatedManagedController_WithGate(t *testing.T) {
	g := new(gate.Gate[schema.GroupVersionKind])
	o := controller.Options{
		Logger: logging.NewNopLogger(),
		Gate:   g,
	}

	err := SetupGatedManagedController(
		nil,
		o,
		iamv1alpha1.PolicyGroupVersionKind,
		iamv1alpha1.PolicyGroupKind,
		&iamv1alpha1.Policy{},
		&DynatraceConnector{},
	)

	if err != nil {
		t.Fatalf("SetupGatedManagedController failed: %v", err)
	}
}
