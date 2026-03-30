package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const experimentsPath = "/console/v1/experiments"

func (c *Client) ListExperiments(ctx context.Context, limit, page int, tags []string) ([]Experiment, *PaginationInfo, error) {
	resp, err := c.list(ctx, experimentsPath, limit, page, tags)
	if err != nil {
		return nil, nil, err
	}
	var experiments []Experiment
	if err := json.Unmarshal(resp.Data, &experiments); err != nil {
		return nil, nil, err
	}
	return experiments, resp.Pagination, nil
}

func (c *Client) GetExperiment(ctx context.Context, id string) (*Experiment, error) {
	raw, err := c.getEntity(ctx, fmt.Sprintf("%s/%s", experimentsPath, id))
	if err != nil {
		return nil, err
	}
	var exp Experiment
	if err := json.Unmarshal(raw, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (c *Client) CreateExperiment(ctx context.Context, name, description string, groups []Group) (*Experiment, error) {
	body := map[string]any{"name": name}
	if description != "" {
		body["description"] = description
	}
	if len(groups) > 0 {
		body["groups"] = groups
	}
	raw, err := c.do(ctx, http.MethodPost, experimentsPath, body)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var exp Experiment
	if err := json.Unmarshal(resp.Data, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (c *Client) DeleteExperiment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", experimentsPath, id), nil)
	return err
}

func (c *Client) ArchiveExperiment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/archive", experimentsPath, id), nil)
	return err
}

func (c *Client) UpdateExperiment(ctx context.Context, id string, update map[string]any) (*Experiment, error) {
	raw, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", experimentsPath, id), update)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var exp Experiment
	if err := json.Unmarshal(resp.Data, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (c *Client) StartExperiment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/start", experimentsPath, id), nil)
	return err
}

func (c *Client) ResetExperiment(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/reset", experimentsPath, id), nil)
	return err
}

func (c *Client) AbandonExperiment(ctx context.Context, id string, reason string) error {
	body := map[string]any{"decisionReason": reason}
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/abandon", experimentsPath, id), body)
	return err
}

func (c *Client) ShipExperiment(ctx context.Context, id string, groupID, reason string, removeTargeting bool) error {
	body := map[string]any{
		"id":               groupID,
		"decisionReason":   reason,
		"removeTargeting":  removeTargeting,
	}
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/make_decision", experimentsPath, id), body)
	return err
}
