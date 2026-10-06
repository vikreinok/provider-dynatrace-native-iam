package dynatrace

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *dynatraceClient) LookupHostEntity(ctx context.Context, entityType, entityName string) (string, []HostEntityTagDto, int, error) {
	var selector string
	if entityName == "*" {
		selector = fmt.Sprintf(`type("%s")`, entityType)
	} else if len(entityName) > 1 && entityName[len(entityName)-1] == '*' {
		prefix := entityName[:len(entityName)-1]
		selector = fmt.Sprintf(`type("%s"),entityName.startsWith("%s")`, entityType, prefix)
	} else {
		selector = fmt.Sprintf(`type("%s"),entityName.equals("%s")`, entityType, entityName)
	}

	path := fmt.Sprintf("/api/v2/entities?entitySelector=%s&fields=tags", url.QueryEscape(selector))
	var out HostEntitiesResponseDto
	if err := c.doEnvRequest(ctx, http.MethodGet, path, nil, &out); err != nil {
		return "", nil, 0, err
	}

	if len(out.Entities) > 0 {
		count := out.TotalCount
		if count == 0 {
			count = len(out.Entities)
		}
		return out.Entities[0].EntityID, out.Entities[0].Tags, count, nil
	}

	return "", nil, 0, nil
}
