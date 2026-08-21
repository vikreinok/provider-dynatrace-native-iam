package dynatrace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) resolvePolicyLevel(levelType, levelID string) (string, string) {
	lt := levelType
	if lt == "" {
		lt = "account"
	}
	lid := levelID
	if lid == "" {
		switch lt {
		case "account":
			lid = c.accountID
		case "global":
			lid = "global"
		}
	}
	return lt, lid
}

func (c *dynatraceClient) ListPolicies(ctx context.Context, levelType, levelID string) (*PolicyListDto, error) {
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/policies", lt, lid)
	var out PolicyListDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetPolicy(ctx context.Context, levelType, levelID, uuid string) (*PolicyDto, error) {
	if uuid == "" {
		return nil, errors.New("policy uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/policies/%s", lt, lid, uuid)
	var out PolicyDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) CreatePolicy(ctx context.Context, levelType, levelID string, policy PolicyDto) (*PolicyDto, error) {
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/policies", lt, lid)
	var out PolicyDto
	if err := c.doRequest(ctx, http.MethodPost, path, policy, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) UpdatePolicy(ctx context.Context, levelType, levelID, uuid string, policy PolicyDto) (*PolicyDto, error) {
	if uuid == "" {
		return nil, errors.New("policy uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/policies/%s", lt, lid, uuid)
	var out PolicyDto
	if err := c.doRequest(ctx, http.MethodPut, path, policy, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) DeletePolicy(ctx context.Context, levelType, levelID, uuid string) error {
	if uuid == "" {
		return errors.New("policy uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/policies/%s", lt, lid, uuid)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
