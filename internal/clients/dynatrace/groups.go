package dynatrace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) ListGroups(ctx context.Context) (*GroupListDto, error) {
	path := fmt.Sprintf("/iam/v1/accounts/%s/groups", c.accountID)
	var out GroupListDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetGroup(ctx context.Context, uuid string) (*GroupDto, error) {
	if uuid == "" {
		return nil, errors.New("group uuid cannot be empty")
	}
	list, err := c.ListGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range list.Items {
		if g.UUID == uuid {
			return &g, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("Group with uuid %s not found in account %s", uuid, c.accountID),
	}
}

func (c *dynatraceClient) CreateGroup(ctx context.Context, group GroupDto) (*GroupDto, error) {
	path := fmt.Sprintf("/iam/v1/accounts/%s/groups", c.accountID)
	payload := []GroupDto{group}
	var out []GroupDto
	if err := c.doRequest(ctx, http.MethodPost, path, payload, &out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("Dynatrace API returned empty group creation response")
	}
	return &out[0], nil
}

func (c *dynatraceClient) UpdateGroup(ctx context.Context, uuid string, group GroupDto) error {
	if uuid == "" {
		return errors.New("group uuid cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/groups/%s", c.accountID, uuid)
	return c.doRequest(ctx, http.MethodPut, path, group, nil)
}

func (c *dynatraceClient) DeleteGroup(ctx context.Context, uuid string) error {
	if uuid == "" {
		return errors.New("group uuid cannot be empty")
	}
	path := fmt.Sprintf("/iam/v1/accounts/%s/groups/%s", c.accountID, uuid)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
