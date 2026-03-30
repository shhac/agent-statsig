package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

func TestFilterConfigs(t *testing.T) {
	configs := []api.DynamicConfig{
		{Name: "feature_flags", Description: "Main feature flags"},
		{Name: "pricing_config", Description: "Pricing tiers"},
		{Name: "ui_settings", Description: "UI customization"},
	}

	result := filterConfigs(configs, "feature")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "feature_flags" {
		t.Errorf("name = %q", result[0].Name)
	}

	result = filterConfigs(configs, "CONFIG")
	if len(result) != 1 {
		t.Errorf("case-insensitive: expected 1, got %d", len(result))
	}
}

func TestValidateAgainstSchema(t *testing.T) {
	schema := json.RawMessage(`{
		"properties": {
			"theme": {"type": "string"},
			"limit": {"type": "number"}
		},
		"required": ["theme"]
	}`)

	// Valid value
	err := validateAgainstSchema(schema, map[string]any{"theme": "dark", "limit": 10})
	if err != nil {
		t.Errorf("valid value should pass: %v", err)
	}

	// Missing required field
	err = validateAgainstSchema(schema, map[string]any{"limit": 10})
	if err == nil {
		t.Error("should error on missing required field")
	}

	// Unknown field
	err = validateAgainstSchema(schema, map[string]any{"theme": "dark", "unknown": true})
	if err == nil {
		t.Error("should error on unknown field")
	}
}

func TestValidateAgainstSchemaNoSchema(t *testing.T) {
	err := validateAgainstSchema(nil, map[string]any{"anything": true})
	if err != nil {
		t.Errorf("nil schema should pass: %v", err)
	}
}

func TestValidateAgainstSchemaNonObject(t *testing.T) {
	schema := json.RawMessage(`{"properties": {"x": {}}}`)
	err := validateAgainstSchema(schema, "not a map")
	if err != nil {
		t.Errorf("non-object value should pass (no validation): %v", err)
	}
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := mapKeys(m)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}
