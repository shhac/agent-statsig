package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

const (
	DefaultBaseURL = "https://statsigapi.net"
	APIVersion     = "20240601"
)

type Client struct {
	baseURL    string
	consoleKey string
	clientKey  string
	http       *http.Client
}

func NewClient(consoleKey, clientKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		consoleKey: consoleKey,
		clientKey:  clientKey,
		http:       &http.Client{},
	}
}

// NewTestClient creates a client pointing at a custom base URL (for tests).
func NewTestClient(baseURL, consoleKey, clientKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		consoleKey: consoleKey,
		clientKey:  clientKey,
		http:       &http.Client{},
	}
}

func (c *Client) HasClientKey() bool {
	return c.clientKey != ""
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	reqURL := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}

	req.Header.Set("STATSIG-API-KEY", c.consoleKey)
	req.Header.Set("STATSIG-API-VERSION", APIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByRetry).WithHint("Network error — check connectivity")
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByRetry)
	}

	if resp.StatusCode >= 400 {
		return nil, classifyHTTPError(resp.StatusCode, respBody)
	}

	return json.RawMessage(respBody), nil
}

// doAndDecode performs an API call and decodes the response's "data" field into T.
func doAndDecode[T any](c *Client, ctx context.Context, method, path string, body any) (*T, error) {
	raw, err := c.do(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	var resp entityResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	var result T
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	return &result, nil
}

// getAndDecode performs a GET and decodes the response's "data" field into T.
func getAndDecode[T any](c *Client, ctx context.Context, path string) (*T, error) {
	return doAndDecode[T](c, ctx, http.MethodGet, path, nil)
}

func classifyHTTPError(status int, body []byte) *agenterrors.APIError {
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	_ = json.Unmarshal(body, &parsed)

	msg := parsed.Message
	if msg == "" {
		msg = parsed.Error
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}

	switch {
	case status == 401:
		return agenterrors.New("Authentication failed: invalid API key", agenterrors.FixableByHuman).
			WithHint("Check your console key with 'agent-statsig project test'")
	case status == 403:
		return agenterrors.New("Permission denied: "+msg, agenterrors.FixableByHuman).
			WithHint("Your API key may not have sufficient permissions")
	case status == 404:
		return agenterrors.New("Not found: "+msg, agenterrors.FixableByAgent).
			WithHint("Check the entity name — use 'list' to see available items")
	case status == 429:
		return agenterrors.New("Rate limited", agenterrors.FixableByRetry).
			WithHint("Statsig rate limit: ~100 requests/10s. Wait and retry.")
	case status >= 500:
		return agenterrors.New("Statsig API error: "+msg, agenterrors.FixableByRetry).
			WithHint("Statsig server error — retry in a few seconds")
	default:
		return agenterrors.New(msg, agenterrors.FixableByAgent)
	}
}

type listResponse struct {
	Data       json.RawMessage `json:"data"`
	Pagination *PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	ItemsPerPage int    `json:"itemsPerPage"`
	PageNumber   int    `json:"pageNumber"`
	TotalItems   int    `json:"totalItems"`
	NextPage     string `json:"nextPage"`
	PreviousPage string `json:"previousPage"`
	All          string `json:"all"`
}

func (p *PaginationInfo) HasMore() bool {
	return p != nil && p.NextPage != ""
}

func (c *Client) list(ctx context.Context, path string, limit, page int, tags []string) (*listResponse, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	for _, tag := range tags {
		params.Add("tags", tag)
	}
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	raw, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp listResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	return &resp, nil
}

type entityResponse struct {
	Data json.RawMessage `json:"data"`
}
