//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

func getLiveClient(t *testing.T) dtclient.Client {
	accountID := os.Getenv("DT_ACCOUNT_ID")
	clientID := os.Getenv("DT_CLIENT_ID")
	clientSecret := os.Getenv("DT_CLIENT_SECRET")
	envURL := os.Getenv("DT_ENV_URL")

	if accountID == "" || clientID == "" || clientSecret == "" {
		// Attempt to read from secret.yaml if present
		if data, err := os.ReadFile("../../secret.yaml"); err == nil {
			type secretDoc struct {
				StringData struct {
					Credentials string `yaml:"credentials"`
				} `yaml:"stringData"`
			}
			var doc secretDoc
			if err := yaml.Unmarshal(data, &doc); err == nil && doc.StringData.Credentials != "" {
				creds, err := dtclient.ParseCredentialsJSON([]byte(doc.StringData.Credentials))
				if err == nil {
					accountID = creds.AccountID
					clientID = creds.ClientID
					clientSecret = creds.ClientSecret
					envURL = creds.EnvURL
				}
			}
		}
	}

	if accountID == "" || clientID == "" || clientSecret == "" {
		t.Skip("Live integration test skipped: DT_ACCOUNT_ID, DT_CLIENT_ID, and DT_CLIENT_SECRET environment variables or secret.yaml required")
	}

	c, err := dtclient.NewClient(dtclient.Credentials{
		AccountID:    accountID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		EnvURL:       envURL,
	})
	if err != nil {
		t.Fatalf("failed to create live Dynatrace client: %v", err)
	}
	return c
}

func TestLive_GroupLifecycle(t *testing.T) {
	client := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	groupName := "E2E Test Group Native " + time.Now().Format("150405")
	t.Logf("Creating live group: %s", groupName)

	// 1. Create
	created, err := client.CreateGroup(ctx, dtclient.GroupDto{
		Name:                     groupName,
		Description:              "Created by native provider E2E integration test",
		FederatedAttributeValues: []string{"e2e-test-saml"},
	})
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	t.Logf("Group created with UUID: %s", created.UUID)

	defer func() {
		t.Logf("Cleaning up group: %s", created.UUID)
		_ = client.DeleteGroup(context.Background(), created.UUID)
	}()

	// 2. Get / Observe
	grp, err := client.GetGroup(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}
	if grp.Name != groupName {
		t.Errorf("GetGroup name = %s, want %s", grp.Name, groupName)
	}

	// 3. Update
	updatedDesc := "Updated description at " + time.Now().Format(time.RFC3339)
	err = client.UpdateGroup(ctx, created.UUID, dtclient.GroupDto{
		Name:                     groupName,
		Description:              updatedDesc,
		FederatedAttributeValues: []string{"e2e-test-saml-updated"},
	})
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	// 4. Verify Update
	grpUpdated, err := client.GetGroup(ctx, created.UUID)
	if err != nil {
		t.Fatalf("GetGroup after update failed: %v", err)
	}
	if grpUpdated.Description != updatedDesc {
		t.Errorf("GetGroup description = %s, want %s", grpUpdated.Description, updatedDesc)
	}

	// 5. Delete
	err = client.DeleteGroup(ctx, created.UUID)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	// 6. Verify Deletion
	_, err = client.GetGroup(ctx, created.UUID)
	if err == nil {
		t.Errorf("expected 404 error after deletion, got nil")
	}
	t.Log("Group lifecycle verified successfully on live Dynatrace API")
}

func TestLive_CostCenterLifecycle(t *testing.T) {
	client := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ccKey := "E2E-CC-" + time.Now().Format("150405")
	t.Logf("Adding live cost center: %s", ccKey)

	// 1. Add
	err := client.AddCostCenter(ctx, ccKey)
	if err != nil {
		t.Fatalf("AddCostCenter failed: %v", err)
	}

	defer func() {
		t.Logf("Cleaning up cost center: %s", ccKey)
		_ = client.DeleteCostCenter(context.Background(), ccKey)
	}()

	// 2. Get / Observe
	cc, err := client.GetCostCenter(ctx, ccKey)
	if err != nil {
		t.Fatalf("GetCostCenter failed: %v", err)
	}
	if cc.Key != ccKey {
		t.Errorf("GetCostCenter key = %s, want %s", cc.Key, ccKey)
	}

	// 3. Delete
	err = client.DeleteCostCenter(ctx, ccKey)
	if err != nil {
		t.Fatalf("DeleteCostCenter failed: %v", err)
	}

	// 4. Verify Deletion
	_, err = client.GetCostCenter(ctx, ccKey)
	if err == nil {
		t.Errorf("expected error getting deleted cost center, got nil")
	}
	t.Log("CostCenter lifecycle verified successfully on live Dynatrace API")
}

func TestLive_PolicyAndBindingsLifecycle(t *testing.T) {
	client := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Create temporary group
	groupName := "E2E Test Bindings Group " + time.Now().Format("150405")
	grp, err := client.CreateGroup(ctx, dtclient.GroupDto{
		Name:        groupName,
		Description: "For bindings E2E test",
	})
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer func() {
		_ = client.DeleteGroup(context.Background(), grp.UUID)
	}()

	// 2. Create policy
	polName := "E2E Test Policy " + time.Now().Format("150405")
	createdPol, err := client.CreatePolicy(ctx, "account", "", dtclient.PolicyDto{
		Name:           polName,
		Description:    "E2E test policy for bindings",
		StatementQuery: "ALLOW settings:objects:read;",
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}
	defer func() {
		_ = client.DeletePolicy(context.Background(), "account", "", createdPol.UUID)
	}()

	// 3. Set policy binding
	err = client.SetPolicyBinding(ctx, "account", "", createdPol.UUID, grp.UUID, dtclient.AppendLevelPolicyBindingForGroupDto{
		Parameters: map[string]string{"env": "test"},
	})
	if err != nil {
		t.Fatalf("SetPolicyBinding failed: %v", err)
	}

	// 4. Get policy bindings for group
	bindings, err := client.GetPolicyBindingsForGroup(ctx, "account", "", grp.UUID)
	if err != nil {
		t.Fatalf("GetPolicyBindingsForGroup failed: %v", err)
	}
	found := false
	for _, b := range bindings.PolicyBindings {
		if b.ID == createdPol.UUID {
			found = true
			break
		}
	}
	if !found {
		for _, u := range bindings.PolicyUUIDs {
			if u == createdPol.UUID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("expected policy %s in group bindings, got: %+v", createdPol.UUID, bindings)
	}

	// 5. Delete policy binding
	err = client.DeletePolicyBinding(ctx, "account", "", createdPol.UUID, grp.UUID)
	if err != nil {
		t.Fatalf("DeletePolicyBinding failed: %v", err)
	}
	t.Log("Policy and PolicyBindings lifecycle verified successfully on live Dynatrace API")
}

