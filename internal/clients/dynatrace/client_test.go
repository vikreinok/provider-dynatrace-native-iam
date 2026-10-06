package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildEnvRequest_AuthorizationPrecedence(t *testing.T) {
	ctx := context.Background()

	t.Run("PlatformTokenTakesPrecedenceOverApiToken", func(t *testing.T) {
		c := &dynatraceClient{
			envURL:        "https://test.live.dynatrace.com",
			platformToken: "dt0s16.SERVICE_USER_TOKEN",
			apiToken:      "dt0c01.PERSONAL_API_TOKEN",
		}

		req, err := c.buildEnvRequest(ctx, http.MethodGet, "https://test.live.dynatrace.com/api/v2/settings/objects", nil)
		if err != nil {
			t.Fatalf("buildEnvRequest failed: %v", err)
		}

		wantAuth := "Bearer dt0s16.SERVICE_USER_TOKEN"
		if got := req.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("expected Authorization %q, got %q", wantAuth, got)
		}
	})

	t.Run("FallsBackToApiTokenWhenPlatformTokenEmpty", func(t *testing.T) {
		c := &dynatraceClient{
			envURL:   "https://test.live.dynatrace.com",
			apiToken: "dt0c01.PERSONAL_API_TOKEN",
		}

		req, err := c.buildEnvRequest(ctx, http.MethodGet, "https://test.live.dynatrace.com/api/v2/settings/objects", nil)
		if err != nil {
			t.Fatalf("buildEnvRequest failed: %v", err)
		}

		wantAuth := "Api-Token dt0c01.PERSONAL_API_TOKEN"
		if got := req.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("expected Authorization %q, got %q", wantAuth, got)
		}
	})

	t.Run("FallsBackToTokenManagerWhenBothTokensEmpty", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintln(w, `{"access_token":"mock-oauth-token","expires_in":3600}`)
		}))
		defer srv.Close()

		tm := NewTokenManager("test-client", "test-secret", WithTokenURL(srv.URL))
		c := &dynatraceClient{
			envURL:       "https://test.live.dynatrace.com",
			tokenManager: tm,
		}

		req, err := c.buildEnvRequest(ctx, http.MethodGet, "https://test.live.dynatrace.com/api/v2/settings/objects", nil)
		if err != nil {
			t.Fatalf("buildEnvRequest failed: %v", err)
		}

		wantAuth := "Bearer mock-oauth-token"
		if got := req.Header.Get("Authorization"); got != wantAuth {
			t.Errorf("expected Authorization %q, got %q", wantAuth, got)
		}
	})

	t.Run("ErrorsWhenNoCredentialsAvailable", func(t *testing.T) {
		c := &dynatraceClient{
			envURL: "https://test.live.dynatrace.com",
		}

		_, err := c.buildEnvRequest(ctx, http.MethodGet, "https://test.live.dynatrace.com/api/v2/settings/objects", nil)
		if err == nil {
			t.Fatalf("expected error when no credentials configured, got nil")
		}
	})
}

func TestBuildRequest_NilTokenManager(t *testing.T) {
	c := &dynatraceClient{
		baseURL: "https://api.dynatrace.com",
	}

	_, err := c.buildRequest(context.Background(), http.MethodGet, "https://api.dynatrace.com/iam/v1/accounts/123/groups", nil)
	if err == nil {
		t.Fatalf("expected error when tokenManager is nil, got nil")
	}
}
