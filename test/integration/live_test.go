//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	dtclient "github.com/vikreinok/provider-dynatrace-native-iam/internal/clients/dynatrace"
)

func getLiveCredentials(t *testing.T) dtclient.Credentials {
	accountID := os.Getenv("DT_ACCOUNT_ID")
	clientID := os.Getenv("DT_CLIENT_ID")
	clientSecret := os.Getenv("DT_CLIENT_SECRET")
	envURL := os.Getenv("DT_ENV_URL")
	apiToken := os.Getenv("DT_API_TOKEN")

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
					apiToken = creds.APIToken
				}
			}
		}
	}

	if accountID == "" || clientID == "" || clientSecret == "" {
		t.Skip("Live integration test skipped: DT_ACCOUNT_ID, DT_CLIENT_ID, and DT_CLIENT_SECRET environment variables or secret.yaml required")
	}

	return dtclient.Credentials{
		AccountID:    accountID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		EnvURL:       envURL,
		APIToken:     apiToken,
	}
}


func getLiveClient(t *testing.T) (dtclient.Client, dtclient.Credentials) {
	creds := getLiveCredentials(t)
	c, err := dtclient.NewClient(creds)
	if err != nil {
		t.Fatalf("failed to create live Dynatrace client: %v", err)
	}
	return c, creds
}

// -----------------------------------------------------------------------------
// SCENARIO 1: Group Lifecycle & Edge Cases
// -----------------------------------------------------------------------------

// [Positive Test] Full Group CRUD lifecycle on live Dynatrace API.
func TestLive_GroupLifecycle(t *testing.T) {
	client, _ := getLiveClient(t)
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

// [Negative Test] Attempting to create a group with invalid empty name fails with API error.
func TestLive_Group_InvalidParams(t *testing.T) {
	client, _ := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.CreateGroup(ctx, dtclient.GroupDto{
		Name: "",
	})
	if err == nil {
		t.Errorf("expected error creating group with empty name, got nil")
	} else {
		t.Logf("Correctly received error for empty group name: %v", err)
	}
}

// -----------------------------------------------------------------------------
// SCENARIO 2: Policy Lifecycle & Negative Query Test
// -----------------------------------------------------------------------------

// [Positive Test] Full Policy CRUD lifecycle on live Dynatrace API.
func TestLive_PolicyLifecycle(t *testing.T) {
	client, creds := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	policyName := "E2E Test Policy Native " + time.Now().Format("150405")
	t.Logf("Creating live policy: %s", policyName)

	// 1. Create Policy
	created, err := client.CreatePolicy(ctx, "account", creds.AccountID, dtclient.PolicyDto{
		Name:           policyName,
		Description:    "E2E policy test",
		StatementQuery: "ALLOW settings:objects:read;",
		Tags:           []string{"e2e:test"},
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}
	t.Logf("Policy created with UUID: %s", created.UUID)

	defer func() {
		t.Logf("Cleaning up policy: %s", created.UUID)
		_ = client.DeletePolicy(context.Background(), "account", creds.AccountID, created.UUID)
	}()

	// 2. Get / Observe Policy
	pol, err := client.GetPolicy(ctx, "account", creds.AccountID, created.UUID)
	if err != nil {
		t.Fatalf("GetPolicy failed: %v", err)
	}
	if pol.Name != policyName {
		t.Errorf("GetPolicy name = %s, want %s", pol.Name, policyName)
	}

	// 3. Update Policy
	updatedDesc := "Updated policy at " + time.Now().Format(time.RFC3339)
	_, err = client.UpdatePolicy(ctx, "account", creds.AccountID, created.UUID, dtclient.PolicyDto{
		Name:           policyName,
		Description:    updatedDesc,
		StatementQuery: "ALLOW settings:objects:read, settings:schemas:read;",
		Tags:           []string{"e2e:test", "e2e:updated"},
	})
	if err != nil {
		t.Fatalf("UpdatePolicy failed: %v", err)
	}

	// 4. Verify Update
	polUpdated, err := client.GetPolicy(ctx, "account", creds.AccountID, created.UUID)
	if err != nil {
		t.Fatalf("GetPolicy after update failed: %v", err)
	}
	if polUpdated.Description != updatedDesc {
		t.Errorf("GetPolicy description = %s, want %s", polUpdated.Description, updatedDesc)
	}

	// 5. Delete Policy
	err = client.DeletePolicy(ctx, "account", creds.AccountID, created.UUID)
	if err != nil {
		t.Fatalf("DeletePolicy failed: %v", err)
	}

	// 6. Verify Deletion
	_, err = client.GetPolicy(ctx, "account", creds.AccountID, created.UUID)
	if err == nil {
		t.Errorf("expected error getting deleted policy, got nil")
	}
	t.Log("Policy lifecycle verified successfully on live Dynatrace API")
}

// [Negative Test] Creating a Policy with invalid DQL syntax fails with API error.
func TestLive_Policy_InvalidStatement(t *testing.T) {
	client, creds := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.CreatePolicy(ctx, "account", creds.AccountID, dtclient.PolicyDto{
		Name:           "Invalid DQL Policy " + time.Now().Format("150405"),
		StatementQuery: "INVALID_SYNTAX_NOT_A_VALID_DQL_STATEMENT;;;",
	})
	if err == nil {
		t.Errorf("expected error for invalid DQL statement query, got nil")
	} else {
		t.Logf("Correctly rejected invalid policy syntax: %v", err)
	}
}

// -----------------------------------------------------------------------------
// SCENARIO 3: Policy Boundary Lifecycle
// -----------------------------------------------------------------------------

// [Positive Test] Full PolicyBoundary CRUD lifecycle on live Dynatrace API.
func TestLive_PolicyBoundaryLifecycle(t *testing.T) {
	client, creds := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	boundaryName := "E2E Test Boundary " + time.Now().Format("150405")
	t.Logf("Creating live boundary: %s", boundaryName)

	// 1. Create Boundary
	created, err := client.CreateBoundary(ctx, "account", creds.AccountID, dtclient.PolicyBoundaryDto{
		Name:          boundaryName,
		BoundaryQuery: "environment:management-zone = 'Production';",
	})
	if err != nil {
		t.Fatalf("CreateBoundary failed: %v", err)
	}
	t.Logf("Boundary created with UUID: %s", created.UUID)

	defer func() {
		t.Logf("Cleaning up boundary: %s", created.UUID)
		_ = client.DeleteBoundary(context.Background(), "account", creds.AccountID, created.UUID)
	}()

	// 2. Get / Observe Boundary
	bnd, err := client.GetBoundary(ctx, "account", creds.AccountID, created.UUID)
	if err != nil {
		t.Fatalf("GetBoundary failed: %v", err)
	}
	if bnd.Name != boundaryName {
		t.Errorf("GetBoundary name = %s, want %s", bnd.Name, boundaryName)
	}

	// 3. Delete Boundary
	err = client.DeleteBoundary(ctx, "account", creds.AccountID, created.UUID)
	if err != nil {
		t.Fatalf("DeleteBoundary failed: %v", err)
	}
	t.Log("PolicyBoundary lifecycle verified successfully on live Dynatrace API")
}

// -----------------------------------------------------------------------------
// SCENARIO 4: Policy Bindings Lifecycle (Positive & Negative)
// -----------------------------------------------------------------------------

// [Positive Test] Policy Binding to Group lifecycle on live Dynatrace API.
func TestLive_PolicyBindingsLifecycle(t *testing.T) {
	client, creds := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Step 1: Create prerequisite Group
	grp, err := client.CreateGroup(ctx, dtclient.GroupDto{
		Name: "E2E Binding Group " + time.Now().Format("150405"),
	})
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer func() { _ = client.DeleteGroup(context.Background(), grp.UUID) }()

	// Step 2: Create prerequisite Policy
	pol, err := client.CreatePolicy(ctx, "account", creds.AccountID, dtclient.PolicyDto{
		Name:           "E2E Binding Policy " + time.Now().Format("150405"),
		StatementQuery: "ALLOW settings:objects:read;",
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}
	defer func() { _ = client.DeletePolicy(context.Background(), "account", creds.AccountID, pol.UUID) }()

	// Step 3: Create Binding
	err = client.SetPolicyBinding(ctx, "account", creds.AccountID, pol.UUID, grp.UUID, dtclient.AppendLevelPolicyBindingForGroupDto{
		Parameters: map[string]string{},
	})
	if err != nil {
		t.Fatalf("SetPolicyBinding failed: %v", err)
	}

	// Step 4: Verify Binding exists
	binding, err := client.GetPolicyBinding(ctx, "account", creds.AccountID, pol.UUID, grp.UUID)
	if err != nil {
		t.Fatalf("GetPolicyBinding failed: %v", err)
	}
	if binding == nil {
		t.Fatalf("expected non-nil policy binding")
	}

	// Step 5: Delete Binding
	err = client.DeletePolicyBinding(ctx, "account", creds.AccountID, pol.UUID, grp.UUID)
	if err != nil {
		t.Fatalf("DeletePolicyBinding failed: %v", err)
	}
	t.Log("PolicyBindings lifecycle verified successfully on live Dynatrace API")
}

// [Negative Test] Binding with non-existent Group/Policy UUID fails with error.
func TestLive_PolicyBindings_InvalidPolicyOrGroup(t *testing.T) {
	client, creds := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.SetPolicyBinding(ctx, "account", creds.AccountID, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000", dtclient.AppendLevelPolicyBindingForGroupDto{
		Parameters: map[string]string{},
	})
	if err == nil {
		t.Errorf("expected error binding non-existent policy to non-existent group, got nil")
	} else {
		t.Logf("Correctly rejected invalid policy binding: %v", err)
	}
}

func TestLive_PolicyAndBindingsLifecycle(t *testing.T) {
	client, creds := getLiveClient(t)
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
	createdPol, err := client.CreatePolicy(ctx, "account", creds.AccountID, dtclient.PolicyDto{
		Name:           polName,
		Description:    "E2E test policy for bindings",
		StatementQuery: "ALLOW settings:objects:read;",
	})
	if err != nil {
		t.Fatalf("CreatePolicy failed: %v", err)
	}
	defer func() {
		_ = client.DeletePolicy(context.Background(), "account", creds.AccountID, createdPol.UUID)
	}()

	// 3. Set policy binding
	err = client.SetPolicyBinding(ctx, "account", creds.AccountID, createdPol.UUID, grp.UUID, dtclient.AppendLevelPolicyBindingForGroupDto{
		Parameters: map[string]string{"env": "test"},
	})
	if err != nil {
		t.Fatalf("SetPolicyBinding failed: %v", err)
	}

	// 4. Get policy bindings for group
	bindings, err := client.GetPolicyBindingsForGroup(ctx, "account", creds.AccountID, grp.UUID)
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
	err = client.DeletePolicyBinding(ctx, "account", creds.AccountID, createdPol.UUID, grp.UUID)
	if err != nil {
		t.Fatalf("DeletePolicyBinding failed: %v", err)
	}
	t.Log("Policy and PolicyBindings lifecycle verified successfully on live Dynatrace API")
}

// -----------------------------------------------------------------------------
// SCENARIO 5: Cost Center Lifecycle
// -----------------------------------------------------------------------------

// [Positive Test] CostCenter Add/Get/Delete lifecycle on live Dynatrace API.
func TestLive_CostCenterLifecycle(t *testing.T) {
	client, _ := getLiveClient(t)
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

// [Positive Test] CostCenter duplicate creation handles already-exists error gracefully.
func TestLive_CostCenterAlreadyExists(t *testing.T) {
	client, _ := getLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ccKey := "E2E-CC-DUP-" + time.Now().Format("150405")
	t.Logf("Creating initial cost center: %s", ccKey)

	// 1. Initial creation
	err := client.AddCostCenter(ctx, ccKey)
	if err != nil {
		t.Fatalf("Initial AddCostCenter failed: %v", err)
	}

	defer func() {
		t.Logf("Cleaning up cost center: %s", ccKey)
		_ = client.DeleteCostCenter(context.Background(), ccKey)
	}()

	// 2. Add same cost center again -> should return already exists error
	dupErr := client.AddCostCenter(ctx, ccKey)
	if dupErr == nil {
		t.Fatalf("Expected duplicate AddCostCenter to return error, got nil")
	}
	if !dtclient.IsAlreadyExists(dupErr) {
		t.Fatalf("Expected duplicate error to be identified by IsAlreadyExists, got: %v", dupErr)
	}
	t.Logf("Successfully verified duplicate error is recognized by IsAlreadyExists: %v", dupErr)

	// 3. Verify GetCostCenter finds the pre-existing entry
	cc, err := client.GetCostCenter(ctx, ccKey)
	if err != nil {
		t.Fatalf("GetCostCenter failed for existing cost center: %v", err)
	}
	if cc.Key != ccKey {
		t.Errorf("GetCostCenter key = %s, want %s", cc.Key, ccKey)
	}
}

// Helper to seed a cost center directly via API for Kind cluster adoption test
func TestLive_SeedPreexistingCostCenter(t *testing.T) {
	client, _ := getLiveClient(t)
	ctx := context.Background()
	key := "E2E-PREEXISTING-API-CC"
	err := client.AddCostCenter(ctx, key)
	if err != nil && !dtclient.IsAlreadyExists(err) {
		t.Fatalf("Failed to seed cost center: %v", err)
	}
	t.Logf("Seeded pre-existing cost center in Dynatrace API: %s", key)
}

// -----------------------------------------------------------------------------
// SCENARIO 6: Authentication & Credentials Failure
// -----------------------------------------------------------------------------

// [Negative Test] Invalid client secret fails OAuth token exchange.
func TestLive_InvalidCredentials(t *testing.T) {
	creds := getLiveCredentials(t)
	invalidCreds := dtclient.Credentials{
		AccountID:    creds.AccountID,
		ClientID:     creds.ClientID,
		ClientSecret: "DEFINITELY_WRONG_CLIENT_SECRET",
		EnvURL:       creds.EnvURL,
	}

	c, err := dtclient.NewClient(invalidCreds)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = c.ListGroups(ctx)
	if err == nil {
		t.Errorf("expected OAuth authentication failure with wrong client secret, got nil")
	} else {
		if !strings.Contains(err.Error(), "OAuth") && !strings.Contains(err.Error(), "token") && !strings.Contains(err.Error(), "401") && !strings.Contains(err.Error(), "400") {
			t.Logf("Received expected error: %v", err)
		} else {
			t.Logf("Correctly received OAuth authentication failure: %v", err)
		}
	}
}
