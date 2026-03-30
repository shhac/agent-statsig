package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewTestClient(srv.URL, "test-console-key", "test-client-key")
}

func entityJSON(data any) []byte {
	b, _ := json.Marshal(map[string]any{"data": data})
	return b
}

func listJSON(data any, total int) []byte {
	b, _ := json.Marshal(map[string]any{
		"data": data,
		"pagination": map[string]any{
			"itemsPerPage": 100,
			"pageNumber":   1,
			"totalItems":   total,
		},
	})
	return b
}

func TestNewClient(t *testing.T) {
	c := NewClient("console-key", "client-key")
	if c.consoleKey != "console-key" {
		t.Errorf("consoleKey = %q", c.consoleKey)
	}
	if !c.HasClientKey() {
		t.Error("HasClientKey() should be true")
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestNewClientNoClientKey(t *testing.T) {
	c := NewClient("console-key", "")
	if c.HasClientKey() {
		t.Error("HasClientKey() should be false with empty client key")
	}
}

func TestDoSendsHeaders(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("STATSIG-API-KEY") != "test-console-key" {
			t.Errorf("API key header = %q", r.Header.Get("STATSIG-API-KEY"))
		}
		if r.Header.Get("STATSIG-API-VERSION") != APIVersion {
			t.Errorf("API version header = %q", r.Header.Get("STATSIG-API-VERSION"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		w.Write(entityJSON(map[string]any{}))
	})

	_, err := client.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatalf("do() error: %v", err)
	}
}

func TestDoWithBody(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("body name = %v", body["name"])
		}
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(map[string]any{"id": "123"}))
	})

	_, err := client.do(context.Background(), "POST", "/test", map[string]any{"name": "test"})
	if err != nil {
		t.Fatalf("do() error: %v", err)
	}
}

func TestDoWithCancelledContext(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(map[string]any{}))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.do(ctx, "GET", "/test", nil)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestClassifyHTTPError401(t *testing.T) {
	err := classifyHTTPError(401, []byte(`{"message":"Unauthorized"}`))
	if err.FixableBy != agenterrors.FixableByHuman {
		t.Errorf("401 should be fixable by human, got %q", err.FixableBy)
	}
}

func TestClassifyHTTPError404(t *testing.T) {
	err := classifyHTTPError(404, []byte(`{"message":"Gate not found"}`))
	if err.FixableBy != agenterrors.FixableByAgent {
		t.Errorf("404 should be fixable by agent, got %q", err.FixableBy)
	}
}

func TestClassifyHTTPError429(t *testing.T) {
	err := classifyHTTPError(429, []byte(`{}`))
	if err.FixableBy != agenterrors.FixableByRetry {
		t.Errorf("429 should be fixable by retry, got %q", err.FixableBy)
	}
}

func TestClassifyHTTPError500(t *testing.T) {
	err := classifyHTTPError(500, []byte(`{"message":"Internal error"}`))
	if err.FixableBy != agenterrors.FixableByRetry {
		t.Errorf("500 should be fixable by retry, got %q", err.FixableBy)
	}
}

func TestClassifyHTTPError403(t *testing.T) {
	err := classifyHTTPError(403, []byte(`{"message":"Forbidden"}`))
	if err.FixableBy != agenterrors.FixableByHuman {
		t.Errorf("403 should be fixable by human, got %q", err.FixableBy)
	}
}

func TestClassifyHTTPErrorEmptyBody(t *testing.T) {
	err := classifyHTTPError(418, []byte(`{}`))
	if err.Message != "HTTP 418" {
		t.Errorf("message = %q, want 'HTTP 418'", err.Message)
	}
}

func TestPaginationHasMore(t *testing.T) {
	p := &PaginationInfo{NextPage: "/console/v1/gates?page=2"}
	if !p.HasMore() {
		t.Error("HasMore() should be true when NextPage is set")
	}

	p2 := &PaginationInfo{}
	if p2.HasMore() {
		t.Error("HasMore() should be false when NextPage is empty")
	}

	var nilP *PaginationInfo
	if nilP.HasMore() {
		t.Error("HasMore() should be false on nil")
	}
}

func TestListQueryParams(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "10" {
			t.Errorf("limit = %q", q.Get("limit"))
		}
		if q.Get("page") != "2" {
			t.Errorf("page = %q", q.Get("page"))
		}
		tags := q["tags"]
		if len(tags) != 2 || tags[0] != "core" || tags[1] != "mobile" {
			t.Errorf("tags = %v", tags)
		}
		w.Write(listJSON([]any{}, 0))
	})

	client.list(context.Background(), "/console/v1/gates", 10, 2, []string{"core", "mobile"})
}

func TestListNoParams(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		w.Write(listJSON([]any{}, 0))
	})

	client.list(context.Background(), "/console/v1/gates", 0, 0, nil)
}

// Gate API tests

func TestGetGate(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/my_gate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(Gate{ID: "gate-123", Name: "my_gate", IsEnabled: true}))
	})

	gate, err := client.GetGate(context.Background(), "my_gate")
	if err != nil {
		t.Fatal(err)
	}
	if gate.Name != "my_gate" {
		t.Errorf("name = %q", gate.Name)
	}
	if !gate.IsEnabled {
		t.Error("isEnabled should be true")
	}
}

func TestCreateGate(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new_gate" {
			t.Errorf("name = %v", body["name"])
		}
		w.Write(entityJSON(Gate{ID: "g-1", Name: "new_gate"}))
	})

	gate, err := client.CreateGate(context.Background(), "new_gate", "A gate")
	if err != nil {
		t.Fatal(err)
	}
	if gate.Name != "new_gate" {
		t.Errorf("name = %q", gate.Name)
	}
}

func TestDeleteGate(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write([]byte(`{"message":"ok"}`))
	})

	if err := client.DeleteGate(context.Background(), "my_gate"); err != nil {
		t.Fatal(err)
	}
}

func TestEnableGate(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/my_gate/enable" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write([]byte(`{"message":"ok"}`))
	})

	if err := client.EnableGate(context.Background(), "my_gate"); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateGate(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(Gate{ID: "g-1", Name: "updated"}))
	})

	gate, err := client.UpdateGate(context.Background(), "g-1", map[string]any{"description": "new"})
	if err != nil {
		t.Fatal(err)
	}
	if gate.Name != "updated" {
		t.Errorf("name = %q", gate.Name)
	}
}

func TestListGates(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(listJSON([]Gate{{Name: "g1"}, {Name: "g2"}}, 2))
	})

	gates, pag, err := client.ListGates(context.Background(), 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(gates) != 2 {
		t.Errorf("got %d gates", len(gates))
	}
	if pag.TotalItems != 2 {
		t.Errorf("totalItems = %d", pag.TotalItems)
	}
}

func TestHTTPError(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Gate not found"}`))
	})

	_, err := client.GetGate(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *agenterrors.APIError
	if !agenterrors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.FixableBy != agenterrors.FixableByAgent {
		t.Errorf("fixableBy = %q", apiErr.FixableBy)
	}
}

// Config API tests

func TestGetConfig(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(DynamicConfig{Name: "my_config"}))
	})

	cfg, err := client.GetConfig(context.Background(), "my_config")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "my_config" {
		t.Errorf("name = %q", cfg.Name)
	}
}

// Experiment API tests

func TestGetExperiment(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(Experiment{Name: "my_exp", Status: "active"}))
	})

	exp, err := client.GetExperiment(context.Background(), "my_exp")
	if err != nil {
		t.Fatal(err)
	}
	if exp.Status != "active" {
		t.Errorf("status = %q", exp.Status)
	}
}

func TestStartExperiment(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/experiments/exp1/start" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"message":"ok"}`))
	})

	if err := client.StartExperiment(context.Background(), "exp1"); err != nil {
		t.Fatal(err)
	}
}

func TestShipExperiment(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["id"] != "group-1" {
			t.Errorf("group id = %v", body["id"])
		}
		if body["decisionReason"] != "winner" {
			t.Errorf("reason = %v", body["decisionReason"])
		}
		w.Write([]byte(`{"message":"ok"}`))
	})

	if err := client.ShipExperiment(context.Background(), "exp1", "group-1", "winner", false); err != nil {
		t.Fatal(err)
	}
}

// Segment API tests

func TestGetSegment(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(Segment{Name: "internal_team", Type: "id_list"}))
	})

	seg, err := client.GetSegment(context.Background(), "internal_team")
	if err != nil {
		t.Fatal(err)
	}
	if seg.Type != "id_list" {
		t.Errorf("type = %q", seg.Type)
	}
}

func TestAddSegmentIDs(t *testing.T) {
	client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		ids := body["ids"].([]any)
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
		w.Write([]byte(`{"message":"ok"}`))
	})

	if err := client.AddSegmentIDs(context.Background(), "seg1", []string{"user1", "user2"}); err != nil {
		t.Fatal(err)
	}
}
