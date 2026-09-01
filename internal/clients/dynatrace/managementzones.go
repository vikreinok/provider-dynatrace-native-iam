package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const managementZoneSchemaID = "builtin:management-zones"

func (c *dynatraceClient) ListManagementZonesV2(ctx context.Context) (*SettingsObjectsListDto, error) {
	path := fmt.Sprintf("/api/v2/settings/objects?schemaIds=%s&fields=objectId,schemaId,scope,value&pageSize=500", url.QueryEscape(managementZoneSchemaID))
	var out SettingsObjectsListDto
	if err := c.doEnvRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) GetManagementZoneV2(ctx context.Context, objectID string) (*SettingsObjectItemDto, error) {
	path := fmt.Sprintf("/api/v2/settings/objects/%s", url.PathEscape(objectID))
	var out SettingsObjectItemDto
	if err := c.doEnvRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *dynatraceClient) CreateManagementZoneV2(ctx context.Context, value ManagementZoneV2Value) (*SettingsObjectResponseDto, error) {
	path := "/api/v2/settings/objects"
	payload := []SettingsObjectCreateDto{
		{
			SchemaID: managementZoneSchemaID,
			Scope:    "environment",
			Value:    value,
		},
	}
	var out []SettingsObjectResponseDto
	if err := c.doEnvRequest(ctx, http.MethodPost, path, payload, &out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("empty response from Dynatrace settings objects create API")
	}
	if out[0].Code >= 400 {
		return nil, fmt.Errorf("failed to create management zone: %s (code %d)", out[0].Message, out[0].Code)
	}
	return &out[0], nil
}

func (c *dynatraceClient) UpdateManagementZoneV2(ctx context.Context, objectID string, value ManagementZoneV2Value) error {
	path := fmt.Sprintf("/api/v2/settings/objects/%s", url.PathEscape(objectID))
	payload := SettingsObjectUpdateDto{
		Value: value,
	}
	return c.doEnvRequest(ctx, http.MethodPut, path, payload, nil)
}

func (c *dynatraceClient) DeleteManagementZoneV2(ctx context.Context, objectID string) error {
	path := fmt.Sprintf("/api/v2/settings/objects/%s", url.PathEscape(objectID))
	return c.doEnvRequest(ctx, http.MethodDelete, path, nil, nil)
}
