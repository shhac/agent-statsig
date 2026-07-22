// Package mockstatsig provides realistic Statsig Console API fixtures and a
// stateful config server for httptest-based CLI tests.
//
// Responses mirror the real API's shapes. Notably, the API stores a dynamic
// config's JSON Schema string-form (JSON-encoded text, not a nested object) —
// object-form fixtures mask bugs in schema handling, so tests should build
// schemas with StringFormSchema.
package mockstatsig

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/shhac/agent-statsig/internal/api"
)

// Entity wraps data in the Console API single-entity envelope.
func Entity(data any) []byte {
	b, _ := json.Marshal(map[string]any{"data": data})
	return b
}

// List wraps data in the Console API list envelope with pagination.
func List(data any, total int) []byte {
	b, _ := json.Marshal(map[string]any{
		"data":       data,
		"pagination": map[string]any{"itemsPerPage": 100, "pageNumber": 1, "totalItems": total},
	})
	return b
}

// StringFormSchema encodes an object-form JSON Schema the way the real API
// stores it: as a JSON string.
func StringFormSchema(schemaJSON string) json.RawMessage {
	b, _ := json.Marshal(schemaJSON)
	return json.RawMessage(b)
}

// ConfigServer serves a single dynamic config with GET/PATCH semantics,
// applying patches to its document and recording every patch body so tests
// can assert on what the CLI actually sent.
type ConfigServer struct {
	mu            sync.Mutex
	doc           map[string]any
	patches       []map[string]any
	rulePatches   []map[string]any
	dryRunPatches []map[string]any
	ruleDeletes   []string
}

func NewConfigServer(cfg api.DynamicConfig) *ConfigServer {
	b, _ := json.Marshal(cfg)
	var doc map[string]any
	_ = json.Unmarshal(b, &doc)
	return &ConfigServer{doc: doc}
}

func (s *ConfigServer) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		switch {
		case r.Method == http.MethodGet:
			_, _ = w.Write(Entity(s.doc))
		case r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/rules/"):
			var patch map[string]any
			_ = json.NewDecoder(r.Body).Decode(&patch)
			s.rulePatches = append(s.rulePatches, patch)
			_, _ = w.Write([]byte(`{"message":"ok"}`))
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/rules/"):
			s.ruleDeletes = append(s.ruleDeletes, path.Base(r.URL.Path))
			_, _ = w.Write([]byte(`{"message":"ok"}`))
		case r.Method == http.MethodPatch:
			var patch map[string]any
			_ = json.NewDecoder(r.Body).Decode(&patch)
			if r.URL.Query().Get("dryRun") == "true" {
				s.dryRunPatches = append(s.dryRunPatches, patch)
				_, _ = w.Write(Entity(s.doc))
				return
			}
			s.patches = append(s.patches, patch)
			for k, v := range patch {
				if v == nil {
					delete(s.doc, k)
					continue
				}
				s.doc[k] = v
			}
			_, _ = w.Write(Entity(s.doc))
		default:
			http.Error(w, `{"message":"unsupported method in mockstatsig.ConfigServer"}`, http.StatusMethodNotAllowed)
		}
	}
}

// Patches returns a copy of the config-level PATCH bodies received, in order.
func (s *ConfigServer) Patches() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]map[string]any(nil), s.patches...)
}

// RulePatches returns a copy of the per-rule PATCH bodies received, in order.
func (s *ConfigServer) RulePatches() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]map[string]any(nil), s.rulePatches...)
}

// DryRunPatches returns a copy of the config-level PATCH bodies that arrived
// with dryRun=true; those are recorded but never applied to the document.
func (s *ConfigServer) DryRunPatches() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]map[string]any(nil), s.dryRunPatches...)
}

// RuleDeletes returns the rule IDs deleted via DELETE /rules/{id}, in order.
func (s *ConfigServer) RuleDeletes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.ruleDeletes...)
}
