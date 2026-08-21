package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) ListUsers(ctx context.Context) (*UserListDto, error) {
	path := fmt.Sprintf("/iam/v1/accounts/%s/users", c.accountID)
	var out UserListDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetUser(ctx context.Context, email string) (*UserDto, error) {
	if email == "" {
		return nil, errors.New("user email cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/users/%s", c.accountID, url.PathEscape(email))
	var out UserDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) CreateUser(ctx context.Context, email string, groups []string) error {
	if email == "" {
		return errors.New("user email cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/users", c.accountID)
	payload := UserEmailDto{Email: email, Groups: groups}
	return c.doRequest(ctx, http.MethodPost, path, payload, nil)
}

func (c *dynatraceClient) DeleteUser(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("user email cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/users/%s", c.accountID, url.PathEscape(email))
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

func (c *dynatraceClient) ListServiceUsers(ctx context.Context) (*ServiceUserListDto, error) {
	path := fmt.Sprintf("/iam/v1/accounts/%s/service-users", c.accountID)
	var out ServiceUserListDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetServiceUser(ctx context.Context, userUUID string) (*ServiceUserDto, error) {
	if userUUID == "" {
		return nil, errors.New("service user uuid cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/service-users/%s", c.accountID, userUUID)
	var out ServiceUserDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) CreateServiceUser(ctx context.Context, user ServiceUserDto) (*ServiceUserDto, error) {
	path := fmt.Sprintf("/iam/v1/accounts/%s/service-users", c.accountID)
	var out ServiceUserDto
	if err := c.doRequest(ctx, http.MethodPost, path, user, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) DeleteServiceUser(ctx context.Context, userUUID string) error {
	if userUUID == "" {
		return errors.New("service user uuid cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/service-users/%s", c.accountID, userUUID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
