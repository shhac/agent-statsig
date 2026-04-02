package gate

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func globals() *shared.GlobalFlags {
	return &shared.GlobalFlags{}
}

func runGateCmd(t *testing.T, handler http.HandlerFunc, args ...string) (string, string) {
	t.Helper()
	shared.SetupMockServer(t, handler)

	root := &cobra.Command{Use: "test"}
	Register(root, func() *shared.GlobalFlags { return globals() })

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	root.SetArgs(args)
	root.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(rOut)
	errBuf.ReadFrom(rErr)

	return outBuf.String(), errBuf.String()
}

func entityJSON(data any) []byte {
	b, _ := json.Marshal(map[string]any{"data": data})
	return b
}

func listJSON(data any, total int) []byte {
	b, _ := json.Marshal(map[string]any{
		"data":       data,
		"pagination": map[string]any{"itemsPerPage": 100, "pageNumber": 1, "totalItems": total},
	})
	return b
}

func TestGateList(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write(listJSON([]api.Gate{{Name: "gate1"}, {Name: "gate2"}}, 2))
	}, "gate", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateGet(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/gates/my_gate" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write(entityJSON(api.Gate{Name: "my_gate", IsEnabled: true}))
	}, "gate", "get", "my_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "my_gate" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestGateCreate(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new_gate" {
			t.Errorf("name = %v", body["name"])
		}
		w.Write(entityJSON(api.Gate{Name: "new_gate"}))
	}, "gate", "create", "new_gate", "--description", "A test gate")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateDelete(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "gate", "disable", "my_gate")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["isEnabled"] != false {
		t.Errorf("isEnabled = %v", parsed["isEnabled"])
	}
}

func TestGateUpdate(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(api.Gate{Name: "updated"}))
	}, "gate", "update", "my_gate", `{"description":"new"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "updated" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestGateUpdateInvalidJSON(t *testing.T) {
	_, stderr := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.Method {
		case "GET":
			w.Write(entityJSON(api.Gate{Name: "my_gate", Rules: []api.Rule{}}))
		case "POST":
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			w.Write(entityJSON(api.Rule{Name: "Everyone"}))
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(entityJSON(api.Gate{
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON([]api.Rule{{ID: "r1", Name: "Email rule"}}))
	}, "gate", "rule", "list", "my_gate")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateRuleAdd(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Write(entityJSON(api.Rule{ID: "new-rule", Name: "Team"}))
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
	_, stderr := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
	root := &cobra.Command{Use: "test"}
	Register(root, func() *shared.GlobalFlags { return globals() })

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	root.SetArgs([]string{"gate", "criteria"})
	root.Execute()

	wOut.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(rOut)

	var parsed map[string]any
	json.Unmarshal(buf.Bytes(), &parsed)
	criteria := parsed["criteria"].([]any)
	if len(criteria) != 25 {
		t.Errorf("expected 25 criteria, got %d", len(criteria))
	}
}

func TestGateListWithSearch(t *testing.T) {
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(listJSON([]api.Gate{
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/console/v1/tags" {
			w.Write(listJSON([]api.Tag{{Name: "mobile"}}, 1))
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		tags := body["tags"].([]any)
		if len(tags) != 1 || tags[0] != "mobile" {
			t.Errorf("tags = %v", tags)
		}
		w.Write(entityJSON(api.Gate{Name: "my_gate", Tags: []string{"mobile"}}))
	}, "gate", "create", "my_gate", "--tag", "mobile")

	if out == "" {
		t.Error("expected output")
	}
}

func TestGateCreateWithInvalidTag(t *testing.T) {
	_, stderr := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/console/v1/tags" {
			w.Write(listJSON([]api.Tag{{Name: "existing"}}, 1))
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
	out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
