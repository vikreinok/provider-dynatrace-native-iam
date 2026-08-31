package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"

	xpv2 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vikreinok/provider-dynatrace-native-iam/apis/v1alpha1"
)

// RawCredentials is used to parse various credential formats found in Dynatrace provider secrets.
type RawCredentials struct {
	AccountID      string `json:"iam_account_id"`
	DTAccountID    string `json:"dt_account_id"`
	ClientID       string `json:"iam_client_id"`
	DTClientID     string `json:"dt_client_id"`
	ClientSecret   string `json:"iam_client_secret"`
	DTClientSecret string `json:"dt_client_secret"`
	EnvURL         string `json:"dt_env_url"`
	APIToken       string `json:"dt_api_token"`
}

// GetClientFromProviderConfig retrieves credentials from a ClusterProviderConfig or ProviderConfig and returns a Dynatrace Client.
func GetClientFromProviderConfig(ctx context.Context, kube client.Client, pcr *xpv2.Reference) (Client, error) {
	pcName := "default"
	if pcr != nil && pcr.Name != "" {
		pcName = pcr.Name
	}

	// Try ClusterProviderConfig first
	cpc := &v1alpha1.ClusterProviderConfig{}
	err := kube.Get(ctx, types.NamespacedName{Name: pcName}, cpc)
	if err == nil {
		return clientFromCredentials(ctx, kube, cpc.Spec.Credentials)
	}

	return nil, errors.Wrapf(err, "cannot find ClusterProviderConfig %s", pcName)
}

// ParseCredentialsJSON parses raw JSON credential data into Credentials.
func ParseCredentialsJSON(data []byte) (Credentials, error) {
	var raw RawCredentials
	if err := json.Unmarshal(data, &raw); err != nil {
		return Credentials{}, errors.Wrap(err, "failed to unmarshal JSON credentials")
	}

	resolvedAccountID := raw.AccountID
	if resolvedAccountID == "" {
		resolvedAccountID = raw.DTAccountID
	}
	resolvedClientID := raw.ClientID
	if resolvedClientID == "" {
		resolvedClientID = raw.DTClientID
	}
	resolvedClientSecret := raw.ClientSecret
	if resolvedClientSecret == "" {
		resolvedClientSecret = raw.DTClientSecret
	}

	return Credentials{
		AccountID:    resolvedAccountID,
		ClientID:     resolvedClientID,
		ClientSecret: resolvedClientSecret,
		EnvURL:       raw.EnvURL,
		APIToken:     raw.APIToken,
	}, nil
}

func clientFromCredentials(ctx context.Context, kube client.Client, creds v1alpha1.ProviderCredentials) (Client, error) {
	if creds.Source != xpv2.CredentialsSourceSecret {
		return nil, fmt.Errorf("unsupported credentials source: %s; only Secret is currently supported", creds.Source)
	}

	if creds.SecretRef == nil {
		return nil, errors.New("missing secretRef in provider credentials")
	}

	secret := &corev1.Secret{}
	err := kube.Get(ctx, types.NamespacedName{
		Namespace: creds.SecretRef.Namespace,
		Name:      creds.SecretRef.Name,
	}, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get credentials secret %s/%s", creds.SecretRef.Namespace, creds.SecretRef.Name)
	}

	key := creds.SecretRef.Key
	if key == "" {
		key = "credentials"
	}

	data, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in credentials secret %s/%s", key, creds.SecretRef.Namespace, creds.SecretRef.Name)
	}

	parsedCreds, err := ParseCredentialsJSON(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse credentials JSON from secret")
	}

	return NewClient(parsedCreds)
}
