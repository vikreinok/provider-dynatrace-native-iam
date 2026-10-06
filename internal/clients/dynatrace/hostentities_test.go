package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupHostEntity(t *testing.T) {
	cases := []struct {
		name             string
		entityType       string
		entityName       string
		platformToken    string
		apiToken         string
		simulate403OnBearer bool
		expectedAuth     string
		responseBody     string
		wantID           string
		wantCount        int
		wantTags         int
		wantErr          bool
	}{
		{
			name:          "ExactMatchPlatformToken",
			entityType:    "HOST",
			entityName:    "CityX",
			platformToken: "dt0s16.TEST_PLATFORM_TOKEN",
			apiToken:      "dt0c01.TEST_API_TOKEN",
			expectedAuth:  "Bearer dt0s16.TEST_PLATFORM_TOKEN",
			responseBody:  `{"totalCount":1,"entities":[{"entityId":"HOST-D6E60CF7996C2E61","type":"HOST","displayName":"CityX","tags":[{"context":"ENVIRONMENT","key":"service","value":"SV-XYZ2"}]}]}`,
			wantID:        "HOST-D6E60CF7996C2E61",
			wantCount:     1,
			wantTags:      1,
		},
		{
			name:                "FallbackToApiTokenWhenPlatformTokenForbidden",
			entityType:          "HOST",
			entityName:          "CityX",
			platformToken:       "dt0s16.NO_SCOPE_TOKEN",
			apiToken:            "dt0c01.TEST_API_TOKEN",
			simulate403OnBearer: true,
			expectedAuth:        "Api-Token dt0c01.TEST_API_TOKEN",
			responseBody:        `{"totalCount":1,"entities":[{"entityId":"HOST-D6E60CF7996C2E61","type":"HOST","displayName":"CityX","tags":[]}]}`,
			wantID:              "HOST-D6E60CF7996C2E61",
			wantCount:           1,
			wantTags:            0,
		},
		{
			name:          "PrefixMatchWithApiToken",
			entityType:    "HOST",
			entityName:    "City*",
			platformToken: "",
			apiToken:      "dt0c01.TEST_API_TOKEN",
			expectedAuth:  "Api-Token dt0c01.TEST_API_TOKEN",
			responseBody:  `{"totalCount":2,"entities":[{"entityId":"HOST-1","type":"HOST","displayName":"City1","tags":[]},{"entityId":"HOST-2","type":"HOST","displayName":"City2","tags":[]}]}`,
			wantID:        "HOST-1",
			wantCount:     2,
			wantTags:      0,
		},
		{
			name:          "WildcardMatchAll",
			entityType:    "HOST",
			entityName:    "*",
			platformToken: "",
			apiToken:      "dt0c01.TEST_API_TOKEN",
			expectedAuth:  "Api-Token dt0c01.TEST_API_TOKEN",
			responseBody:  `{"totalCount":1,"entities":[{"entityId":"HOST-1","type":"HOST","displayName":"HostA","tags":[]}]}`,
			wantID:        "HOST-1",
			wantCount:     1,
			wantTags:      0,
		},
		{
			name:          "NotFoundEmptyResponse",
			entityType:    "HOST",
			entityName:    "nonexistent",
			platformToken: "dt0s16.TEST",
			expectedAuth:  "Bearer dt0s16.TEST",
			responseBody:  `{"totalCount":0,"entities":[]}`,
			wantID:        "",
			wantCount:     0,
			wantTags:      0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var lastAuthHeader string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				lastAuthHeader = r.Header.Get("Authorization")
				if tc.simulate403OnBearer && r.Header.Get("Authorization") == fmt.Sprintf("Bearer %s", tc.platformToken) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = fmt.Fprintln(w, `{"error":{"code":403,"message":"OAuth token is missing required scope"}}`)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintln(w, tc.responseBody)
			}))
			defer server.Close()

			c := &dynatraceClient{
				envURL:        server.URL,
				platformToken: tc.platformToken,
				apiToken:      tc.apiToken,
				httpClient:    server.Client(),
			}

			id, tags, count, err := c.LookupHostEntity(context.Background(), tc.entityType, tc.entityName)
			if (err != nil) != tc.wantErr {
				t.Fatalf("LookupHostEntity error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if lastAuthHeader != tc.expectedAuth {
				t.Errorf("expected Auth header %q, got %q", tc.expectedAuth, lastAuthHeader)
			}
			if id != tc.wantID {
				t.Errorf("expected entity ID %q, got %q", tc.wantID, id)
			}
			if count != tc.wantCount {
				t.Errorf("expected count %d, got %d", tc.wantCount, count)
			}
			if len(tags) != tc.wantTags {
				t.Errorf("expected %d tags, got %d", tc.wantTags, len(tags))
			}
		})
	}
}
