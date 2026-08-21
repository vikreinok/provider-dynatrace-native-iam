package dynatrace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) GetPolicyBindingsForGroup(ctx context.Context, levelType, levelID, groupUUID string) (*PolicyBindingsDto, error) {
	if groupUUID == "" {
		return nil, errors.New("group uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/bindings/groups/%s", lt, lid, groupUUID)
	var out PolicyBindingsDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetPolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) (*PolicyBindingsDto, error) {
	if policyUUID == "" || groupUUID == "" {
		return nil, errors.New("policy uuid and group uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/bindings/%s/%s", lt, lid, policyUUID, groupUUID)
	var out PolicyBindingsDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) SetPolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string, binding AppendLevelPolicyBindingForGroupDto) error {
	if policyUUID == "" || groupUUID == "" {
		return errors.New("policy uuid and group uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/bindings/%s/%s", lt, lid, policyUUID, groupUUID)
	err := c.doRequest(ctx, http.MethodPut, path, binding, nil)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}

func (c *dynatraceClient) DeletePolicyBinding(ctx context.Context, levelType, levelID, policyUUID, groupUUID string) error {
	if policyUUID == "" || groupUUID == "" {
		return errors.New("policy uuid and group uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/bindings/%s/%s", lt, lid, policyUUID, groupUUID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
