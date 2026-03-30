package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const segmentsPath = "/console/v1/segments"

func (c *Client) ListSegments(ctx context.Context, limit, page int, tags []string) ([]Segment, *PaginationInfo, error) {
	resp, err := c.list(ctx, segmentsPath, limit, page, tags)
	if err != nil {
		return nil, nil, err
	}
	var segments []Segment
	if err := json.Unmarshal(resp.Data, &segments); err != nil {
		return nil, nil, err
	}
	return segments, resp.Pagination, nil
}

func (c *Client) GetSegment(ctx context.Context, id string) (*Segment, error) {
	raw, err := c.getEntity(ctx, fmt.Sprintf("%s/%s", segmentsPath, id))
	if err != nil {
		return nil, err
	}
	var seg Segment
	if err := json.Unmarshal(raw, &seg); err != nil {
		return nil, err
	}
	return &seg, nil
}

func (c *Client) CreateSegment(ctx context.Context, name, description, segType string) (*Segment, error) {
	body := map[string]any{"name": name}
	if description != "" {
		body["description"] = description
	}
	if segType != "" {
		body["type"] = segType
	}
	raw, err := c.do(ctx, http.MethodPost, segmentsPath, body)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var seg Segment
	if err := json.Unmarshal(resp.Data, &seg); err != nil {
		return nil, err
	}
	return &seg, nil
}

func (c *Client) DeleteSegment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", segmentsPath, id), nil)
	return err
}

func (c *Client) ArchiveSegment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/archive", segmentsPath, id), nil)
	return err
}

func (c *Client) UpdateSegmentRules(ctx context.Context, id string, rules []Rule) error {
	body := map[string]any{"rules": rules}
	_, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%s/rules", segmentsPath, id), body)
	return err
}

func (c *Client) AddSegmentIDs(ctx context.Context, id string, ids []string) error {
	body := map[string]any{"ids": ids}
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("%s/%s/ids", segmentsPath, id), body)
	return err
}

func (c *Client) RemoveSegmentIDs(ctx context.Context, id string, ids []string) error {
	body := map[string]any{"ids": ids}
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s/ids", segmentsPath, id), body)
	return err
}

func (c *Client) GetSegmentIDs(ctx context.Context, id string) (json.RawMessage, error) {
	return c.getEntity(ctx, fmt.Sprintf("%s/%s/ids", segmentsPath, id))
}
