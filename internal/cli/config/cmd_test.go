package config

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/clitest"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestConfigList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.List([]api.DynamicConfig{{Name: "config1"}, {Name: "config2"}}, 2))
	}, "config", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigGet(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity(api.DynamicConfig{Name: "my_config", IsEnabled: true}))
	}, "config", "get", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "my_config" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestConfigCreate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(mockstatsig.Entity(api.DynamicConfig{Name: "new_config"}))
	}, "config", "create", "new_config")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigDelete(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "config", "delete", "old_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestConfigEnable(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "config", "enable", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["isEnabled"] != true {
		t.Errorf("isEnabled = %v", parsed["isEnabled"])
	}
}

func TestConfigUpdate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity(api.DynamicConfig{Name: "updated"}))
	}, "config", "update", "my_config", `{"description":"new"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "updated" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestConfigUpdateInvalidJSON(t *testing.T) {
	_, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server")
	}, "config", "update", "my_config", "bad-json")

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestConfigRuleList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity([]api.Rule{{ID: "r1", Name: "Default"}}))
	}, "config", "rule", "list", "my_config")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigRuleAddWithSchemaValidation(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(mockstatsig.Entity(api.DynamicConfig{
				Name:   "my_config",
				Schema: mockstatsig.StringFormSchema(`{"properties":{"theme":{"type":"string"}},"required":["theme"]}`),
			}))
		case "PATCH":
			w.Write(mockstatsig.Entity(api.DynamicConfig{Name: "my_config"}))
		}
	}, "config", "rule", "add", "my_config",
		"--name", "Dark theme",
		"--criteria", "email",
		"--operator", "any",
		"--value", "user@test.com",
		"--return-value", `{"theme":"dark"}`)

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigRuleAddSchemaViolation(t *testing.T) {
	_, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write(mockstatsig.Entity(api.DynamicConfig{
				Name:   "my_config",
				Schema: mockstatsig.StringFormSchema(`{"properties":{"theme":{"type":"string"}},"required":["theme"]}`),
			}))
			return
		}
		t.Error("should not reach PATCH with schema violation")
	}, "config", "rule", "add", "my_config",
		"--name", "Bad",
		"--criteria", "email",
		"--operator", "any",
		"--value", "user@test.com",
		"--return-value", `{"unknown_field":true}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestConfigRuleRemove(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:  "my_config",
		Rules: []api.Rule{{ID: "r1", Name: "First"}},
	})
	out, stderr := clitest.Run(t, Register, srv.Handler(), "config", "rule", "remove", "my_config", "--rule", "r1")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["deleted"] != "r1" {
		t.Errorf("deleted = %v", parsed["deleted"])
	}
	deletes := srv.RuleDeletes()
	if len(deletes) != 1 || deletes[0] != "r1" {
		t.Errorf("RuleDeletes = %v", deletes)
	}
}

func TestConfigRuleRemoveByName(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:  "my_config",
		Rules: []api.Rule{{ID: "r1", Name: "First"}, {ID: "r2", Name: "Second"}},
	})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "rule", "remove", "my_config", "--rule", "Second")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	deletes := srv.RuleDeletes()
	if len(deletes) != 1 || deletes[0] != "r2" {
		t.Errorf("name should resolve to the rule's ID; RuleDeletes = %v", deletes)
	}
}

func TestConfigRuleRemoveAmbiguousName(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:  "my_config",
		Rules: []api.Rule{{ID: "r1", Name: "Dup"}, {ID: "r2", Name: "Dup"}},
	})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "rule", "remove", "my_config", "--rule", "Dup")

	if stderr == "" {
		t.Fatal("ambiguous name must error")
	}
	for _, id := range []string{"r1", "r2"} {
		if !strings.Contains(stderr, id) {
			t.Errorf("ambiguity error should list candidate ID %s: %s", id, stderr)
		}
	}
	if len(srv.RuleDeletes()) != 0 {
		t.Error("must not delete on ambiguity")
	}
}

func TestConfigRuleByIDSkipsNameResolution(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:  "my_config",
		Rules: []api.Rule{{ID: "r1", Name: "First"}},
	})

	// move: byID rejects a name outright, nothing is written
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "rule", "move", "my_config",
		"--rule", "First", "--by-id", "--position", "top")
	if stderr == "" {
		t.Error("--by-id must not match names")
	}
	if len(srv.Patches()) != 0 {
		t.Error("no PATCH expected")
	}

	// remove: byID skips the lookup fetch and passes the ref straight through
	_, stderr = clitest.Run(t, Register, srv.Handler(), "config", "rule", "remove", "my_config",
		"--rule", "r1", "--by-id")
	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	if deletes := srv.RuleDeletes(); len(deletes) != 1 || deletes[0] != "r1" {
		t.Errorf("RuleDeletes = %v", deletes)
	}
}
