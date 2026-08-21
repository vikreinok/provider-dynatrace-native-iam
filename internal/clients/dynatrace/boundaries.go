package dynatrace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) ListBoundaries(ctx context.Context, levelType, levelID string) (*PolicyBoundaryListDto, error) {
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/boundaries", lt, lid)
	var out PolicyBoundaryListDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetBoundary(ctx context.Context, levelType, levelID, uuid string) (*PolicyBoundaryDto, error) {
	if uuid == "" {
		return nil, errors.New("policy boundary uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/boundaries/%s", lt, lid, uuid)
	var out PolicyBoundaryDto
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) CreateBoundary(ctx context.Context, levelType, levelID string, boundary PolicyBoundaryDto) (*PolicyBoundaryDto, error) {
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/boundaries", lt, lid)
	var out PolicyBoundaryDto
	if err := c.doRequest(ctx, http.MethodPost, path, boundary, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) UpdateBoundary(ctx context.Context, levelType, levelID, uuid string, boundary PolicyBoundaryDto) (*PolicyBoundaryDto, error) {
	if uuid == "" {
		return nil, errors.New("policy boundary uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/boundaries/%s", lt, lid, uuid)
	var out PolicyBoundaryDto
	if err := c.doRequest(ctx, http.MethodPut, path, boundary, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) DeleteBoundary(ctx context.Context, levelType, levelID, uuid string) error {
	if uuid == "" {
		return errors.New("policy boundary uuid cannot be empty")
	}
	lt, lid := c.resolvePolicyLevel(levelType, levelID)
	path := fmt.Sprintf("/iam/v1/repo/%s/%s/boundaries/%s", lt, lid, uuid)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
