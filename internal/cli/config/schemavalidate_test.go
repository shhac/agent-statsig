package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/clitest"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestNormalizeSchema(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		want   string
		wantOK bool
	}{
		{"empty", "", "", false},
		{"whitespace", "   ", "", false},
		{"null literal", "null", "", false},
		{"object form", `{"type":"object"}`, `{"type":"object"}`, true},
		{"string form", `"{\"type\":\"object\"}"`, `{"type":"object"}`, true},
		{"empty string form", `""`, "", false},
		{"quoted null", `"null"`, "", false},
		{"string wrapping non-JSON text", `"not json"`, "not json", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NormalizeSchema(json.RawMessage(tc.in))
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && string(got) != tc.want {
				t.Errorf("normalized = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfigUpdateValidatesRulesPayload(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "update", "my_config",
		`{"rules":[{"name":"r","returnValue":{"theme":123}}]}`)

	if stderr == "" {
		t.Error("rules payload with non-conforming returnValue should be blocked")
	}
	if len(srv.Patches()) != 0 {
		t.Error("should not PATCH")
	}

	_, stderr = clitest.Run(t, Register, srv.Handler(), "config", "update", "my_config",
		`{"rules":[{"name":"r","returnValue":{"theme":"dark"}},{"name":"no-rv"}]}`)

	if stderr != "" {
		t.Errorf("conforming rules payload should pass (nil returnValue skipped): %s", stderr)
	}
	if len(srv.Patches()) != 1 {
		t.Errorf("PatchCount = %d", len(srv.Patches()))
	}
}

func TestConfigUpdateValidatesAgainstIncomingSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	update := fmt.Sprintf(`{"schema": %q, "defaultValue": {"theme":123}}`, themeSchema)
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "update", "my_config", update)

	if stderr == "" {
		t.Error("defaultValue must be validated against the schema being set in the same update")
	}
	if len(srv.Patches()) != 0 {
		t.Error("should not PATCH")
	}

	update = fmt.Sprintf(`{"schema": %q, "defaultValue": {"theme":"dark"}}`, themeSchema)
	_, stderr = clitest.Run(t, Register, srv.Handler(), "config", "update", "my_config", update)

	if stderr != "" {
		t.Errorf("conforming defaultValue should pass against incoming schema: %s", stderr)
	}
	if len(srv.Patches()) != 1 {
		t.Errorf("PatchCount = %d", len(srv.Patches()))
	}
}

func TestConfigUpdateSchemaNullSkipsValidation(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "update", "my_config",
		`{"schema": null, "defaultValue": {"theme":123}}`)

	if stderr != "" {
		t.Errorf("clearing the schema disables validation for the same update: %s", stderr)
	}
	if len(srv.Patches()) != 1 {
		t.Errorf("PatchCount = %d", len(srv.Patches()))
	}
}

func TestValueSetDryRunStillValidatesClientSide(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "value", "set", "my_config", `{"theme":123}`, "--dry-run")

	if stderr == "" {
		t.Error("client-side validation must run before a --dry-run request")
	}
	if len(srv.DryRunPatches()) != 0 {
		t.Error("blocked value must not reach the API even as a dry run")
	}
}

func TestSchemaSetDryRun(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := clitest.Run(t, Register, srv.Handler(), "config", "schema", "set", "my_config", themeSchema, "--dry-run")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	if len(srv.Patches()) != 0 {
		t.Error("dry-run must not persist")
	}
	if len(srv.DryRunPatches()) != 1 {
		t.Errorf("DryRunPatchCount = %d", len(srv.DryRunPatches()))
	}
}

func TestSchemaGetWithoutSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	out, stderr := clitest.Run(t, Register, srv.Handler(), "config", "schema", "get", "my_config")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "my_config" {
		t.Errorf("name = %v", parsed["name"])
	}
	// The output layer prunes nulls, so a config without a schema emits no
	// schema key at all — that absence is the agent-facing contract.
	if v, present := parsed["schema"]; present {
		t.Errorf("schema key should be pruned when absent, got %v", v)
	}
}
