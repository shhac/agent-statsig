package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const configsPath = "/console/v1/dynamic_configs"

func (c *Client) ListConfigs(ctx context.Context, limit, page int, tags []string) ([]DynamicConfig, *PaginationInfo, error) {
	resp, err := c.list(ctx, configsPath, limit, page, tags)
	if err != nil {
		return nil, nil, err
	}
	var configs []DynamicConfig
	if err := json.Unmarshal(resp.Data, &configs); err != nil {
		return nil, nil, err
	}
	return configs, resp.Pagination, nil
}

func (c *Client) GetConfig(ctx context.Context, id string) (*DynamicConfig, error) {
	return getAndDecode[DynamicConfig](c, ctx, fmt.Sprintf("%s/%s", configsPath, id))
}

func (c *Client) CreateConfig(ctx context.Context, name, description string) (*DynamicConfig, error) {
	body := map[string]any{"name": name}
	if description != "" {
		body["description"] = description
	}
	return doAndDecode[DynamicConfig](c, ctx, http.MethodPost, configsPath, body)
}

func (c *Client) DeleteConfig(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", configsPath, id), nil)
	return err
}

func (c *Client) EnableConfig(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/enable", configsPath, id), nil)
	return err
}

func (c *Client) DisableConfig(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/disable", configsPath, id), nil)
	return err
}

func (c *Client) ArchiveConfig(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPut, fmt.Sprintf("%s/%s/archive", configsPath, id), nil)
	return err
}

func (c *Client) UpdateConfig(ctx context.Context, id string, update map[string]any) (*DynamicConfig, error) {
	return doAndDecode[DynamicConfig](c, ctx, http.MethodPatch, fmt.Sprintf("%s/%s", configsPath, id), update)
}

func (c *Client) GetConfigRules(ctx context.Context, id string) ([]Rule, error) {
	raw, err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s/%s/rules", configsPath, id), nil)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}
	var rules []Rule
	if err := json.Unmarshal(resp.Data, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (c *Client) UpdateConfigRule(ctx context.Context, configID, ruleID string, update map[string]any) error {
	_, err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%s/rules/%s", configsPath, configID, ruleID), update)
	return err
}

func (c *Client) DeleteConfigRule(ctx context.Context, configID, ruleID string) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("%s/%s/rules/%s", configsPath, configID, ruleID), nil)
	return err
}
