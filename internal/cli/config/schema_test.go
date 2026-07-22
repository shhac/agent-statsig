package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

const themeSchema = `{"type":"object","properties":{"theme":{"type":"string"}},"required":["theme"],"additionalProperties":false}`

func TestSchemaGetUnwrapsStringForm(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	out, _ := runConfigCmd(t, srv.Handler(), "config", "schema", "get", "my_config")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	schema, ok := parsed["schema"].(map[string]any)
	if !ok {
		t.Fatalf("schema should be object form, got %T: %v", parsed["schema"], parsed["schema"])
	}
	if schema["type"] != "object" {
		t.Errorf("schema.type = %v", schema["type"])
	}
}

func TestSchemaSetHappyPath(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	out, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", themeSchema)

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	if out == "" {
		t.Error("expected output")
	}
	patch := srv.LastPatch()
	if patch == nil {
		t.Fatal("expected a PATCH")
	}
	encoded, ok := patch["schema"].(string)
	if !ok {
		t.Fatalf("schema must be sent string-form, got %T", patch["schema"])
	}
	var roundTrip map[string]any
	if err := json.Unmarshal([]byte(encoded), &roundTrip); err != nil {
		t.Fatalf("schema string is not valid JSON: %v", err)
	}
	if roundTrip["type"] != "object" {
		t.Errorf("round-tripped schema.type = %v", roundTrip["type"])
	}
}

func TestSchemaSetBlockedByNonConformingValues(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:         "my_config",
		DefaultValue: json.RawMessage(`{"theme":123}`),
		Rules: []api.Rule{
			{ID: "r1", Name: "Bad rule", ReturnValue: map[string]any{"unknown": true}},
		},
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", themeSchema)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH when existing values violate the new schema")
	}
}

func TestSchemaSetEmptyObjectDefaultBlockedByRequired(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:         "my_config",
		DefaultValue: json.RawMessage(`{}`),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", themeSchema)

	if stderr == "" {
		t.Error("empty-object defaultValue should fail a schema with required fields")
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH")
	}
}

func TestSchemaSetNullDefaultSkipsValidation(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:         "my_config",
		DefaultValue: json.RawMessage(`null`),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", themeSchema)

	if stderr != "" {
		t.Errorf("absent/null defaultValue should not block schema set: %s", stderr)
	}
	if srv.PatchCount() != 1 {
		t.Errorf("PatchCount = %d", srv.PatchCount())
	}
}

func TestSchemaSetForceOverridesViolations(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:         "my_config",
		DefaultValue: json.RawMessage(`{"theme":123}`),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", themeSchema, "--force")

	if stderr != "" {
		t.Errorf("--force should skip validation: %s", stderr)
	}
	if srv.PatchCount() != 1 {
		t.Errorf("PatchCount = %d", srv.PatchCount())
	}
}

func TestSchemaSetRejectsInvalidSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", `{"type":"nonsense"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH an invalid schema")
	}
}

func TestSchemaSetRejectsOldDrafts(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config",
		`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH a non-2020-12 schema")
	}
}

func TestSchemaSetAccepts202012SchemaURI(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config",
		`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)

	if stderr != "" {
		t.Errorf("2020-12 $schema should be accepted: %s", stderr)
	}
	if srv.PatchCount() != 1 {
		t.Errorf("PatchCount = %d", srv.PatchCount())
	}
}

func TestSchemaSetRejectsNonObject(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "set", "my_config", `"just a string"`)

	if stderr == "" {
		t.Error("non-object schema should be rejected")
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH")
	}
}

func TestSchemaClear(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "schema", "clear", "my_config")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	patch := srv.LastPatch()
	if patch == nil {
		t.Fatal("expected a PATCH")
	}
	if v, present := patch["schema"]; !present || v != nil {
		t.Errorf("expected schema:null in patch, got %v (present=%v)", v, present)
	}
}

// Regression: the real API returns schemas string-form; validation used to
// silently no-op on them.
func TestRuleAddValidatesStringFormSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "rule", "add", "my_config",
		"--name", "Bad",
		"--criteria", "email",
		"--value", "user@test.com",
		"--return-value", `{"unknown_field":true}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("string-form schema must still be enforced; fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH a schema-violating rule")
	}
}

func TestRuleUpdateValidatesReturnValue(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "rule", "update", "my_config",
		"--rule", "r1", "--return-value", `{"theme":123}`)

	if stderr == "" {
		t.Error("non-conforming return value should be blocked")
	}
	if srv.RulePatchCount() != 0 {
		t.Error("should not PATCH the rule")
	}

	_, stderr = runConfigCmd(t, srv.Handler(), "config", "rule", "update", "my_config",
		"--rule", "r1", "--return-value", `{"theme":"dark"}`)

	if stderr != "" {
		t.Errorf("conforming return value should pass: %s", stderr)
	}
	if srv.RulePatchCount() != 1 {
		t.Errorf("RulePatchCount = %d", srv.RulePatchCount())
	}
}

func TestConfigUpdateValidatesDefaultValue(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name:   "my_config",
		Schema: mockstatsig.StringFormSchema(themeSchema),
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "update", "my_config", `{"defaultValue":{"theme":123}}`)

	if stderr == "" {
		t.Error("non-conforming defaultValue should be blocked")
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH")
	}

	_, stderr = runConfigCmd(t, srv.Handler(), "config", "update", "my_config", `{"defaultValue":{"theme":123}}`, "--force")

	if stderr != "" {
		t.Errorf("--force should skip validation: %s", stderr)
	}
	if srv.PatchCount() != 1 {
		t.Errorf("PatchCount = %d", srv.PatchCount())
	}
}

func TestConfigUpdateRejectsObjectFormSchema(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "update", "my_config", `{"schema":{"type":"object"}}`)

	var parsed map[string]any
	json.Unmarshal([]byte(stderr), &parsed)
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v (stderr: %s)", parsed["fixable_by"], stderr)
	}
	if srv.PatchCount() != 0 {
		t.Error("should not PATCH object-form schema the API would reject")
	}
}

func TestValidateAgainstSchemaStringForm(t *testing.T) {
	schema := mockstatsig.StringFormSchema(themeSchema)

	if err := ValidateAgainstSchema(schema, map[string]any{"theme": "dark"}); err != nil {
		t.Errorf("conforming value should pass: %v", err)
	}
	if err := ValidateAgainstSchema(schema, map[string]any{"theme": 123}); err == nil {
		t.Error("non-conforming value should fail against string-form schema")
	}
}

func TestConfigRuleMove(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{
		Name: "my_config",
		Rules: []api.Rule{
			{ID: "r1", Name: "first"}, {ID: "r2", Name: "second"}, {ID: "r3", Name: "third"},
		},
	})
	_, stderr := runConfigCmd(t, srv.Handler(), "config", "rule", "move", "my_config",
		"--rule", "r3", "--position", "1")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	patch := srv.LastPatch()
	rules, ok := patch["rules"].([]any)
	if !ok || len(rules) != 3 {
		t.Fatalf("patched rules = %v", patch["rules"])
	}
	first, _ := rules[0].(map[string]any)
	if first["id"] != "r3" {
		t.Errorf("first rule after move = %v", first["id"])
	}
}
