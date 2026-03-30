package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const gatesPath = "/console/v1/gates"

func (c *Client) ListGates(ctx context.Context, limit, page int, tags []string) ([]Gate, *PaginationInfo, error) {
	resp, err := c.list(ctx, gatesPath, limit, page, tags)
	if err != nil {
		return nil, nil, err
	}
	var gates []Gate
	if err := json.Unmarshal(resp.Data, &gates); err != nil {
		return nil, nil, err
	}
	return gates, resp.Pagination, nil
}

func (c *Client) GetGate(ctx context.Context, id string) (*Gate, error) {
	raw, err := c.getEntity(ctx, fmt.Sprintf("%s/%s", gatesPath, id))
	if err != nil {
		return nil, err
	}
	var gate Gate
	if err := json.Unmarshal(raw, &gate); err != nil {
		return nil, err
	}
	return &gate, nil
}

func (c *Client) CreateGate(ctx context.Context, name, description string) (*Gate, error) {
	body := map[string]any{"name": name}
	if description != "" {
		body["description"] = description
	}
	raw, err := c.do(ctx, http.MethodPost, gatesPath, body)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var gate Gate
	if err := json.Unmarshal(resp.Data, &gate); err != nil {
		return nil, err
	}
	return &gate, nil
}

func (c *Client) DeleteGate(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", gatesPath, id), nil)
	return err
}

func (c *Client) EnableGate(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/enable", gatesPath, id), nil)
	return err
}

func (c *Client) DisableGate(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/disable", gatesPath, id), nil)
	return err
}

func (c *Client) ArchiveGate(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/archive", gatesPath, id), nil)
	return err
}

func (c *Client) LaunchGate(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/launch", gatesPath, id), nil)
	return err
}

func (c *Client) UpdateGate(ctx context.Context, id string, update map[string]any) (*Gate, error) {
	raw, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", gatesPath, id), update)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var gate Gate
	if err := json.Unmarshal(resp.Data, &gate); err != nil {
		return nil, err
	}
	return &gate, nil
}

func (c *Client) GetGateRules(ctx context.Context, id string) ([]Rule, error) {
	raw, err := c.getEntity(ctx, fmt.Sprintf("%s/%s/rules", gatesPath, id))
	if err != nil {
		return nil, err
	}
	var rules []Rule
	if err := json.Unmarshal(raw, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (c *Client) AddGateRule(ctx context.Context, id string, rule Rule) (*Rule, error) {
	raw, err := c.do(ctx, http.MethodPost, fmt.Sprintf("%s/%s/rule", gatesPath, id), rule)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var created Rule
	if err := json.Unmarshal(resp.Data, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *Client) UpdateGateRule(ctx context.Context, gateID, ruleID string, update map[string]any) error {
	_, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%s/rules/%s", gatesPath, gateID, ruleID), update)
	return err
}

func (c *Client) DeleteGateRule(ctx context.Context, gateID, ruleID string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s/rules/%s", gatesPath, gateID, ruleID), nil)
	return err
}
