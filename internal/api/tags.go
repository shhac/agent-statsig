package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const tagsPath = "/console/v1/tags"

func (c *Client) ListTags(ctx context.Context, limit, page int) ([]Tag, *PaginationInfo, error) {
	resp, err := c.list(ctx, tagsPath, limit, page, nil)
	if err != nil {
		return nil, nil, err
	}
	var tags []Tag
	if err := json.Unmarshal(resp.Data, &tags); err != nil {
		return nil, nil, err
	}
	return tags, resp.Pagination, nil
}

func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	return getAndDecode[Tag](c, ctx, fmt.Sprintf("%s/%s", tagsPath, id))
}

func (c *Client) CreateTag(ctx context.Context, name, description string, isCore bool) (*Tag, error) {
	body := map[string]any{"name": name}
	if description != "" {
		body["description"] = description
	}
	if isCore {
		body["isCore"] = true
	}
	return doAndDecode[Tag](c, ctx, http.MethodPost, tagsPath, body)
}

func (c *Client) UpdateTag(ctx context.Context, id string, update map[string]any) (*Tag, error) {
	return doAndDecode[Tag](c, ctx, http.MethodPatch, fmt.Sprintf("%s/%s", tagsPath, id), update)
}

func (c *Client) DeleteTag(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", tagsPath, id), nil)
	return err
}
