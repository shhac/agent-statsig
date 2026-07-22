package gate

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/clitest"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestGateList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write(mockstatsig.List([]api.Gate{{Name: "gate1"}, {Name: "gate2"}}, 2))
	}, "gate", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateGet(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/my_gate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write(mockstatsig.Entity(api.Gate{Name: "my_gate", IsEnabled: true}))
	}, "gate", "get", "my_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "my_gate" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestGateCreate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new_gate" {
			t.Errorf("name = %v", body["name"])
		}
		w.Write(mockstatsig.Entity(api.Gate{Name: "new_gate"}))
	}, "gate", "create", "new_gate", "--description", "A test gate")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateDelete(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "gate", "delete", "old_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestGateEnable(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/my_gate/enable" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "gate", "enable", "my_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["isEnabled"] != true {
		t.Errorf("isEnabled = %v", parsed["isEnabled"])
	}
}

func TestGateDisable(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "gate", "disable", "my_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["isEnabled"] != false {
		t.Errorf("isEnabled = %v", parsed["isEnabled"])
	}
}

func TestGateUpdate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(mockstatsig.Entity(api.Gate{Name: "updated"}))
	}, "gate", "update", "my_gate", `{"description":"new"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "updated" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestGateUpdateInvalidJSON(t *testing.T) {
	_, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server with invalid JSON")
	}, "gate", "update", "my_gate", "not-json")

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestGateRolloutNew(t *testing.T) {
	callCount := 0
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.Method {
		case "GET":
			w.Write(mockstatsig.Entity(api.Gate{Name: "my_gate", Rules: []api.Rule{}}))
		case "POST":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			w.Write(mockstatsig.Entity(api.Rule{Name: "Everyone"}))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	}, "gate", "rollout", "my_gate", "--percent", "50")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["rolloutPercent"] != float64(50) {
		t.Errorf("rolloutPercent = %v", parsed["rolloutPercent"])
	}
}

func TestGateRolloutExisting(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(mockstatsig.Entity(api.Gate{
				Name: "my_gate",
				Rules: []api.Rule{{
					ID:         "r1",
					Name:       "Everyone",
					Conditions: []api.Condition{{Type: "public"}},
				}},
			}))
		case "PATCH":
			w.Write([]byte(`{"message":"ok"}`))
		}
	}, "gate", "rollout", "my_gate", "--percent", "75")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestGateRuleList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity([]api.Rule{{ID: "r1", Name: "Email rule"}}))
	}, "gate", "rule", "list", "my_gate")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateRuleAdd(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Write(mockstatsig.Entity(api.Rule{ID: "new-rule", Name: "Team"}))
			return
		}
		t.Errorf("unexpected method %s", r.Method)
	}, "gate", "rule", "add", "my_gate",
		"--name", "Team",
		"--criteria", "email",
		"--operator", "str_contains_any",
		"--value", "@company.com",
		"--pass-percent", "100")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateRuleAddInvalidCriteria(t *testing.T) {
	_, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server with invalid criteria")
	}, "gate", "rule", "add", "my_gate",
		"--name", "Bad",
		"--criteria", "nonexistent")

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestGateCriteria(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {}, "gate", "criteria")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	criteria := parsed["criteria"].([]any)
	if len(criteria) != 25 {
		t.Errorf("expected 25 criteria, got %d", len(criteria))
	}
}

func TestGateListWithSearch(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.List([]api.Gate{
			{Name: "feature_onboarding", Description: "Onboarding"},
			{Name: "feature_checkout", Description: "Checkout"},
			{Name: "debug_tool", Description: "Debug"},
		}, 3))
	}, "gate", "list", "--search", "feature")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateCreateWithValidTag(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/console/v1/tags" {
			w.Write(mockstatsig.List([]api.Tag{{Name: "mobile"}}, 1))
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		tags := body["tags"].([]any)
		if len(tags) != 1 || tags[0] != "mobile" {
			t.Errorf("tags = %v", tags)
		}
		w.Write(mockstatsig.Entity(api.Gate{Name: "my_gate", Tags: []string{"mobile"}}))
	}, "gate", "create", "my_gate", "--tag", "mobile")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateCreateWithInvalidTag(t *testing.T) {
	_, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/console/v1/tags" {
			w.Write(mockstatsig.List([]api.Tag{{Name: "existing"}}, 1))
			return
		}
		t.Error("should not reach gate create endpoint")
	}, "gate", "create", "my_gate", "--tag", "nonexistent")

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestGateArchive(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/old_gate/archive" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "gate", "archive", "old_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["archived"] != true {
		t.Errorf("archived = %v", parsed["archived"])
	}
}

func TestGateRuleMove(t *testing.T) {
	var patched map[string]any
	out, stderr := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(mockstatsig.Entity(api.Gate{Name: "my_gate", Rules: []api.Rule{
				{ID: "r1", Name: "first"}, {ID: "r2", Name: "second"},
			}}))
		case "PATCH":
			json.NewDecoder(r.Body).Decode(&patched)
			w.Write(mockstatsig.Entity(api.Gate{Name: "my_gate"}))
		}
	}, "gate", "rule", "move", "my_gate", "--rule", "r2", "--position", "top")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	if out == "" {
		t.Error("expected output")
	}
	rules, ok := patched["rules"].([]any)
	if !ok || len(rules) != 2 {
		t.Fatalf("patched rules = %v", patched["rules"])
	}
	first, _ := rules[0].(map[string]any)
	if first["id"] != "r2" {
		t.Errorf("first rule after move = %v", first["id"])
	}
}
