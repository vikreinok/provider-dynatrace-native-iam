package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

func (c *dynatraceClient) ListCostCenters(ctx context.Context) (*PaginatedFieldValueDto, error) {
	path := fmt.Sprintf("/v1/accounts/%s/settings/costcenters", c.accountID)
	var allRecords []FieldValueDto
	page := 1
	for {
		pagePath := path
		if page > 1 {
			pagePath = fmt.Sprintf("%s?page=%d", path, page)
		}
		var out PaginatedFieldValueDto
		if err := c.doRequest(ctx, http.MethodGet, pagePath, nil, &out); err != nil {
			return nil, err
		}
		allRecords = append(allRecords, out.Records...)
		if !out.HasNextPage || len(out.Records) == 0 {
			break
		}
		page++
	}
	return &PaginatedFieldValueDto{
		Records:     allRecords,
		HasNextPage: false,
	}, nil
}

func (c *dynatraceClient) GetCostCenter(ctx context.Context, key string) (*FieldValueDto, error) {
	if key == "" {
		return nil, errors.New("cost center key cannot be empty")
	}
	list, err := c.ListCostCenters(ctx)
	if err != nil {
		return nil, err
	}
	for _, r := range list.Records {
		if r.Key == key {
			return &r, nil
		}
	}
	return nil, &APIError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("Cost center '%s' not found in account %s", key, c.accountID),
	}
}

func (c *dynatraceClient) AddCostCenter(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("cost center key cannot be empty")
	}
	path := fmt.Sprintf("/v1/accounts/%s/settings/costcenters", c.accountID)
	payload := FieldValuesRequestDto{
		Values: []FieldValueDto{{Key: key}},
	}
	return c.doRequest(ctx, http.MethodPost, path, payload, nil)
}

func (c *dynatraceClient) DeleteCostCenter(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("cost center key cannot be empty")
	}
	path := fmt.Sprintf("/v1/accounts/%s/settings/costcenters/%s", c.accountID, url.PathEscape(key))
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
