package dynatrace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultTokenURL = "https://sso.dynatrace.com/sso/oauth2/token"
	defaultScopes   = "account-idm-read account-idm-write account-env-read iam-policies-management account-uac-read account-uac-write"
)

// TokenManager handles OAuth2 token acquisition and in-memory caching.
type TokenManager struct {
	mu           sync.RWMutex
	clientID     string
	clientSecret string
	tokenURL     string
	scopes       string
	cachedToken  *CachedToken
	httpClient   *http.Client
}

// NewTokenManager creates a new OAuth2 TokenManager.
func NewTokenManager(clientID, clientSecret string, opts ...TokenManagerOption) *TokenManager {
	tm := &TokenManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     defaultTokenURL,
		scopes:       defaultScopes,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(tm)
	}
	return tm
}

// TokenManagerOption configures a TokenManager.
type TokenManagerOption func(*TokenManager)

// WithTokenURL overrides the default token URL.
func WithTokenURL(u string) TokenManagerOption {
	return func(tm *TokenManager) {
		if u != "" {
			tm.tokenURL = u
		}
	}
}

// WithScopes overrides default scopes.
func WithScopes(s string) TokenManagerOption {
	return func(tm *TokenManager) {
		if s != "" {
			tm.scopes = s
		}
	}
}

// WithHTTPClient configures custom HTTP client.
func WithHTTPClient(c *http.Client) TokenManagerOption {
	return func(tm *TokenManager) {
		if c != nil {
			tm.httpClient = c
		}
	}
}

// GetToken returns a valid OAuth token, fetching a new one if expired.
func (tm *TokenManager) GetToken(ctx context.Context) (string, error) {
	tm.mu.RLock()
	if tm.cachedToken != nil && time.Now().Before(tm.cachedToken.ExpiresAt) {
		token := tm.cachedToken.Token
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double check after acquiring write lock
	if tm.cachedToken != nil && time.Now().Before(tm.cachedToken.ExpiresAt) {
		return tm.cachedToken.Token, nil
	}

	return tm.fetchToken(ctx)
}

func (tm *TokenManager) fetchToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", tm.clientID)
	data.Set("client_secret", tm.clientSecret)
	data.Set("scope", tm.scopes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tm.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "failed to create token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send OAuth token request to Dynatrace SSO")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read OAuth token response body")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("dynatrace SSO token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal OAuth token response")
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("dynatrace SSO returned empty access token")
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 300
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn-30) * time.Second)

	tm.cachedToken = &CachedToken{
		Token:     tokenResp.AccessToken,
		ExpiresAt: expiresAt,
	}

	return tokenResp.AccessToken, nil
}
