package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestValueGet(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:         "my_config",
		DefaultValue: json.RawMessage(`{"theme":"light"}`),
	})
	out, _ := runConfigCmd(t, srv.Handler(), "config", "value", "get", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	dv, ok := parsed["defaultValue"].(map[string]any)
	if !ok || dv["theme"] != "light" {
		t.Errorf("defaultValue = %v", parsed["defaultValue"])
	}
}

func TestValueSetValidatesAgainstSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "value", "set", "my_config", `{"theme":123}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if len(srv.Patches()) != 0 {
		t.Error("should not PATCH a schema-violating defaultValue")
	}

	_, stderr = runConfigCmd(t, srv.Handler(), "config", "value", "set", "my_config", `{"theme":"dark"}`)
	if stderr != "" {
		t.Errorf("conforming value should pass: %s", stderr)
	}
	patches := srv.Patches()
	if len(patches) == 0 {
		t.Fatal("expected a PATCH")
	}
	patch := patches[len(patches)-1]
	dv, ok := patch["defaultValue"].(map[string]any)
	if !ok || dv["theme"] != "dark" {
		t.Errorf("patched defaultValue = %v", patch["defaultValue"])
	}
}

func TestValueSetForceSkipsValidation(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "value", "set", "my_config", `{"theme":123}`, "--force")

	if stderr != "" {
		t.Errorf("--force should skip validation: %s", stderr)
	}
	if len(srv.Patches()) != 1 {
		t.Errorf("PatchCount = %d", len(srv.Patches()))
	}
}

func TestValueSetRejectsNonObject(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "value", "set", "my_config", `42`)

	if stderr == "" {
		t.Error("non-object defaultValue should be rejected")
	}
	if len(srv.Patches()) != 0 {
		t.Error("should not PATCH")
	}
}
