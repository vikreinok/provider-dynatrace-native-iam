package dynatrace

import (
	"errors"
	"time"
)

// Credentials contains the extracted credentials needed to communicate with Dynatrace APIs.
type Credentials struct {
	AccountID    string `json:"iam_account_id"`
	ClientID     string `json:"iam_client_id"`
	ClientSecret string `json:"iam_client_secret"`
	EnvURL       string `json:"dt_env_url,omitempty"`
	APIToken     string `json:"dt_api_token,omitempty"`
}

// OAuthTokenResponse represents response from sso.dynatrace.com OAuth token endpoint.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Resource    string `json:"resource,omitempty"`
}

// CachedToken holds a cached OAuth token and expiration timestamp.
type CachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// GroupDto represents a Dynatrace IAM Group.
type GroupDto struct {
	UUID                     string    `json:"uuid,omitempty"`
	Name                     string    `json:"name"`
	Description              string    `json:"description,omitempty"`
	Owner                    string    `json:"owner,omitempty"`
	FederatedAttributeValues []string  `json:"federatedAttributeValues,omitempty"`
	Hidden                   bool      `json:"hidden,omitempty"`
	CreatedAt                string    `json:"createdAt,omitempty"`
	UpdatedAt                string    `json:"updatedAt,omitempty"`
	Permissions              []PermDto `json:"permissions,omitempty"`
}

// PermDto represents a permission assigned to a group.
type PermDto struct {
	PermissionName string `json:"permissionName"`
	Scope          string `json:"scope"`
	ScopeType      string `json:"scopeType"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

// GroupListDto represents paginated list of groups.
type GroupListDto struct {
	Count int        `json:"count"`
	Items []GroupDto `json:"items"`
}

// PolicyDto represents a Dynatrace IAM Policy.
type PolicyDto struct {
	UUID           string   `json:"uuid,omitempty"`
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	StatementQuery string   `json:"statementQuery,omitempty"`
	LevelType      string   `json:"levelType,omitempty"`
	LevelID        string   `json:"levelId,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

// PolicyListDto represents paginated policy list response.
type PolicyListDto struct {
	Policies []PolicyDto `json:"policies"`
}

// PolicyBoundaryDto represents a Dynatrace Policy Boundary.
type PolicyBoundaryDto struct {
	UUID          string            `json:"uuid,omitempty"`
	Name          string            `json:"name"`
	BoundaryQuery string            `json:"boundaryQuery,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	LevelType     string            `json:"levelType,omitempty"`
	LevelID       string            `json:"levelId,omitempty"`
}

// PolicyBoundaryListDto represents a list of policy boundaries.
type PolicyBoundaryListDto struct {
	Boundaries []PolicyBoundaryDto `json:"boundaries"`
}

// BindingItem represents a single policy binding entry in PolicyBindingsV2.
type BindingItem struct {
	ID         string            `json:"id"`
	Boundaries []string          `json:"boundaries,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// PolicyBindingsDto represents the list of bound policies for a group or level.
type PolicyBindingsDto struct {
	LevelType      string        `json:"levelType,omitempty"`
	LevelID        string        `json:"levelId,omitempty"`
	PolicyBindings []BindingItem `json:"policyBindings,omitempty"`
	PolicyUUIDs    []string      `json:"policyUuids,omitempty"`
}

// AppendLevelPolicyBindingForGroupDto represents body for binding PUT/POST.
type AppendLevelPolicyBindingForGroupDto struct {
	Parameters map[string]string `json:"parameters,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Boundaries []string          `json:"boundaries,omitempty"`
}

// FieldValueDto represents a single cost center value.
type FieldValueDto struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// FieldValuesRequestDto represents payload to add/replace cost center values.
type FieldValuesRequestDto struct {
	Values []FieldValueDto `json:"values"`
}

// PaginatedFieldValueDto represents response when listing cost centers.
type PaginatedFieldValueDto struct {
	Records     []FieldValueDto `json:"records"`
	HasNextPage bool            `json:"hasNextPage"`
}

// UserGroupRefDto represents a group reference inside a user object.
type UserGroupRefDto struct {
	UUID string `json:"uuid"`
	Name string `json:"name,omitempty"`
}

// UserDto represents a Dynatrace account user.
type UserDto struct {
	UID     string            `json:"uid,omitempty"`
	Email   string            `json:"email"`
	Name    string            `json:"name,omitempty"`
	Surname string            `json:"surname,omitempty"`
	Type    string            `json:"type,omitempty"`
	Groups  []UserGroupRefDto `json:"groups,omitempty"`
}

// UserEmailDto represents payload to create a new user by email and optional groups.
type UserEmailDto struct {
	Email  string   `json:"email"`
	Groups []string `json:"groups,omitempty"`
}

// UserListDto represents response when listing users.
type UserListDto struct {
	Count int       `json:"count"`
	Items []UserDto `json:"items"`
}

// ServiceUserDto represents a Dynatrace service user.
type ServiceUserDto struct {
	UID         string   `json:"uid,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Email       string   `json:"email,omitempty"`
	Groups      []string `json:"groups,omitempty"`
}

// ServiceUserListDto represents list of service users.
type ServiceUserListDto struct {
	Count int              `json:"count"`
	Items []ServiceUserDto `json:"items"`
}

// APIError represents an error response from Dynatrace API.
type APIError struct {
	StatusCode int
	Message    string
	RawBody    string
}

func (e *APIError) Error() string {
	return e.Message
}

// IsNotFound returns true if the error represents an HTTP 404 Not Found from Dynatrace.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		return true
	}
	return false
}

// isHexByte returns true if the byte is a valid hexadecimal character.
func isHexByte(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// IsUUID returns true if the string is formatted as a standard 36-character UUID.
func IsUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !isHexByte(c) {
			return false
		}
	}
	return true
}
