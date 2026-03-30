package config

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

func runConfigCmd(t *testing.T, handler http.HandlerFunc, args ...string) (string, string) {
	t.Helper()
	shared.SetupMockServer(t, handler)

	root := &cobra.Command{Use: "test"}
	Register(root, func() *shared.GlobalFlags { return globals() })

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

func TestConfigList(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(listJSON([]api.DynamicConfig{{Name: "config1"}, {Name: "config2"}}, 2))
	}, "config", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigGet(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(api.DynamicConfig{Name: "my_config", IsEnabled: true}))
	}, "config", "get", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "my_config" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestConfigCreate(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(api.DynamicConfig{Name: "new_config"}))
	}, "config", "create", "new_config")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigDelete(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "config", "delete", "old_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestConfigEnable(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "config", "enable", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["isEnabled"] != true {
		t.Errorf("isEnabled = %v", parsed["isEnabled"])
	}
}

func TestConfigUpdate(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(api.DynamicConfig{Name: "updated"}))
	}, "config", "update", "my_config", `{"description":"new"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "updated" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestConfigUpdateInvalidJSON(t *testing.T) {
	_, stderr := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach server")
	}, "config", "update", "my_config", "bad-json")

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}

func TestConfigRuleList(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON([]api.Rule{{ID: "r1", Name: "Default"}}))
	}, "config", "rule", "list", "my_config")

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigRuleAddWithSchemaValidation(t *testing.T) {
	out, _ := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write(entityJSON(api.DynamicConfig{
				Name:   "my_config",
				Schema: json.RawMessage(`{"properties":{"theme":{"type":"string"}},"required":["theme"]}`),
			}))
		case "PATCH":
			w.Write(entityJSON(api.DynamicConfig{Name: "my_config"}))
		}
	}, "config", "rule", "add", "my_config",
		"--name", "Dark theme",
		"--criteria", "email",
		"--operator", "any",
		"--values", "user@test.com",
		"--return-value", `{"theme":"dark"}`)

	if out == "" {
		t.Error("expected output")
	}
}

func TestConfigRuleAddSchemaViolation(t *testing.T) {
	_, stderr := runConfigCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write(entityJSON(api.DynamicConfig{
				Name:   "my_config",
				Schema: json.RawMessage(`{"properties":{"theme":{"type":"string"}},"required":["theme"]}`),
			}))
			return
		}
		t.Error("should not reach PATCH with schema violation")
	}, "config", "rule", "add", "my_config",
		"--name", "Bad",
		"--criteria", "email",
		"--operator", "any",
		"--values", "user@test.com",
		"--return-value", `{"unknown_field":true}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
}
