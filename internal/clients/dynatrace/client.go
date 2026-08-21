package dynatrace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultBaseURL = "https://api.dynatrace.com"
	maxRetries     = 3
)

// Client is the interface for interacting with Dynatrace IAM and CostCenter APIs.
type Client interface {
	// Account & Group operations
	ListGroups(ctx context.Context) (*GroupListDto, error)
	GetGroup(ctx context.Context, uuid string) (*GroupDto, error)
	CreateGroup(ctx context.Context, group GroupDto) (*GroupDto, error)
	UpdateGroup(ctx context.Context, uuid string, group GroupDto) error
	DeleteGroup(ctx context.Context, uuid string) error

	// Policy operations
	ListPolicies(ctx context.Context, levelType, levelID string) (*PolicyListDto, error)
	GetPolicy(ctx context.Context, levelType, levelID, uuid string) (*PolicyDto, error)
	CreatePolicy(ctx context.Context, levelType, levelID string, policy PolicyDto) (*PolicyDto, error)
	UpdatePolicy(ctx context.Context, levelType, levelID, uuid string, policy PolicyDto) (*PolicyDto, error)
	DeletePolicy(ctx context.Context, levelType, levelID, uuid string) error

	// Policy Boundary operations
	ListBoundaries(ctx context.Context, levelType, levelID string) (*PolicyBoundaryListDto, error)
	GetBoundary(ctx context.Context, levelType, levelID, uuid string) (*PolicyBoundaryDto, error)
	CreateBoundary(ctx context.Context, levelType, levelID string, boundary PolicyBoundaryDto) (*PolicyBoundaryDto, error)
	UpdateBoundary(ctx context.Context, levelType, levelID, uuid string, boundary PolicyBoundaryDto) (*PolicyBoundaryDto, error)
	DeleteBoundary(ctx context.Context, levelType, levelID, uuid string) error

	// Policy Bindings operations
	GetPolicyBindingsForGroup(ctx context.Context, levelType, levelID, groupUUID string) (*PolicyBindingsDto, error)
	GetPolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) (*PolicyBindingsDto, error)
	SetPolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string, binding AppendLevelPolicyBindingForGroupDto) error
	DeletePolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) error

	// Cost Center operations
	ListCostCenters(ctx context.Context) (*PaginatedFieldValueDto, error)
	GetCostCenter(ctx context.Context, key string) (*FieldValueDto, error)
	AddCostCenter(ctx context.Context, key string) error
	DeleteCostCenter(ctx context.Context, key string) error

	// User & ServiceUser operations
	// Users
	ListUsers(ctx context.Context) (*UserListDto, error)
	GetUser(ctx context.Context, email string) (*UserDto, error)
	CreateUser(ctx context.Context, email string, groups []string) error
	DeleteUser(ctx context.Context, email string) error

	ListServiceUsers(ctx context.Context) (*ServiceUserListDto, error)
	GetServiceUser(ctx context.Context, userUUID string) (*ServiceUserDto, error)
	CreateServiceUser(ctx context.Context, user ServiceUserDto) (*ServiceUserDto, error)
	DeleteServiceUser(ctx context.Context, userUUID string) error
}

type dynatraceClient struct {
	accountID    string
	baseURL      string
	tokenManager *TokenManager
	httpClient   *http.Client
}

// NewClient creates a new Dynatrace IAM API Client.
func NewClient(creds Credentials, opts ...ClientOption) (Client, error) {
	if creds.AccountID == "" {
		return nil, errors.New("missing iam_account_id in Dynatrace credentials")
	}
	if creds.ClientID == "" || creds.ClientSecret == "" {
		return nil, errors.New("missing iam_client_id or iam_client_secret in Dynatrace credentials")
	}

	c := &dynatraceClient{
		accountID:    creds.AccountID,
		baseURL:      defaultBaseURL,
		tokenManager: NewTokenManager(creds.ClientID, creds.ClientSecret),
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// ClientOption configures a Dynatrace client.
type ClientOption func(*dynatraceClient)

// WithBaseURL overrides the API base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *dynatraceClient) {
		if u != "" {
			c.baseURL = strings.TrimSuffix(u, "/")
		}
	}
}

// WithCustomHTTPClient overrides the HTTP client.
func WithCustomHTTPClient(hc *http.Client) ClientOption {
	return func(c *dynatraceClient) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

func (c *dynatraceClient) doRequest(ctx context.Context, method, path string, reqBody any, out any) error {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)

	var rawJSON []byte
	if reqBody != nil {
		var err error
		rawJSON, err = json.Marshal(reqBody)
		if err != nil {
			return errors.Wrap(err, "failed to marshal request body")
		}
	}

	resp, err := c.executeWithRetry(ctx, method, fullURL, rawJSON)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Dynatrace API error (HTTP %d): %s", resp.StatusCode, string(respBody)),
			RawBody:    string(respBody),
		}
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return errors.Wrapf(err, "failed to unmarshal Dynatrace API response from %s", fullURL)
		}
	}

	return nil
}

func (c *dynatraceClient) buildRequest(ctx context.Context, method, fullURL string, rawJSON []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if rawJSON != nil {
		bodyReader = bytes.NewReader(rawJSON)
	}

	token, err := c.tokenManager.GetToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain Dynatrace OAuth access token")
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	if rawJSON != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (c *dynatraceClient) executeWithRetry(ctx context.Context, method, fullURL string, rawJSON []byte) (*http.Response, error) {
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := c.buildRequest(ctx, method, fullURL, rawJSON)
		if err != nil {
			return nil, err
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			_ = resp.Body.Close()
			if err := waitRetryAfter(ctx, resp.Header.Get("Retry-After")); err != nil {
				return nil, err
			}
			continue
		}

		break
	}

	if resp == nil {
		if lastErr != nil {
			return nil, errors.Wrapf(lastErr, "request failed after retries: %s %s", method, fullURL)
		}
		return nil, fmt.Errorf("request failed with no response: %s %s", method, fullURL)
	}

	return resp, nil
}

func waitRetryAfter(ctx context.Context, headerVal string) error {
	retryAfter := 1
	if headerVal != "" {
		if s, parseErr := strconv.Atoi(headerVal); parseErr == nil && s > 0 {
			retryAfter = s
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(retryAfter) * time.Second):
		return nil
	}
}
