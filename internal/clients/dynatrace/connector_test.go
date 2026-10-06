package dynatrace

import (
	"testing"
)

func TestParseCredentialsJSON_PlatformToken(t *testing.T) {
	t.Run("PlatformTokenField", func(t *testing.T) {
		data := []byte(`{
			"dt_env_url": "https://test.live.dynatrace.com",
			"iam_account_id": "test-account",
			"iam_client_id": "test-client",
			"iam_client_secret": "test-secret",
			"platform_token": "dt0s16.TEST_PLATFORM_TOKEN",
			"dt_api_token": "dt0c01.TEST_API_TOKEN"
		}`)

		creds, err := ParseCredentialsJSON(data)
		if err != nil {
			t.Fatalf("ParseCredentialsJSON failed: %v", err)
		}

		if creds.PlatformToken != "dt0s16.TEST_PLATFORM_TOKEN" {
			t.Errorf("expected PlatformToken 'dt0s16.TEST_PLATFORM_TOKEN', got '%s'", creds.PlatformToken)
		}
		if creds.APIToken != "dt0c01.TEST_API_TOKEN" {
			t.Errorf("expected APIToken 'dt0c01.TEST_API_TOKEN', got '%s'", creds.APIToken)
		}
		if creds.AccountID != "test-account" {
			t.Errorf("expected AccountID 'test-account', got '%s'", creds.AccountID)
		}
	})

	t.Run("DTPlatformTokenFallback", func(t *testing.T) {
		data := []byte(`{
			"dt_env_url": "https://test.live.dynatrace.com",
			"dt_platform_token": "dt0s16.FALLBACK_PLATFORM_TOKEN"
		}`)

		creds, err := ParseCredentialsJSON(data)
		if err != nil {
			t.Fatalf("ParseCredentialsJSON failed: %v", err)
		}

		if creds.PlatformToken != "dt0s16.FALLBACK_PLATFORM_TOKEN" {
			t.Errorf("expected PlatformToken 'dt0s16.FALLBACK_PLATFORM_TOKEN', got '%s'", creds.PlatformToken)
		}
	})
}
