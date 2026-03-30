package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func TestValidateAgainstSchema(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"enabledGlobally": {"type": "boolean"},
			"allowOrganizers": {"type": "array", "items": {"type": "string"}},
			"denyOrganizers": {"type": "array", "items": {"type": "string"}}
		},
		"required": ["enabledGlobally", "allowOrganizers", "denyOrganizers"]
	}`)

	// Valid value
	err := ValidateAgainstSchema(schema, map[string]any{
		"enabledGlobally": true,
		"allowOrganizers": []any{},
		"denyOrganizers":  []any{},
	})
	if err != nil {
		t.Errorf("valid value should pass: %v", err)
	}

	// Missing required field
	err = ValidateAgainstSchema(schema, map[string]any{
		"enabledGlobally": true,
		"allowOrganizers": []any{},
	})
	if err == nil {
		t.Error("should error on missing required field")
	}

	// Wrong type (string instead of boolean)
	err = ValidateAgainstSchema(schema, map[string]any{
		"enabledGlobally": "not a bool",
		"allowOrganizers": []any{},
		"denyOrganizers":  []any{},
	})
	if err == nil {
		t.Error("should error on wrong type")
	}

	// Wrong array item type (number instead of string)
	err = ValidateAgainstSchema(schema, map[string]any{
		"enabledGlobally": true,
		"allowOrganizers": []any{123},
		"denyOrganizers":  []any{},
	})
	if err == nil {
		t.Error("should error on wrong array item type")
	}
}

func TestValidateAgainstSchemaNoSchema(t *testing.T) {
	err := ValidateAgainstSchema(nil, map[string]any{"anything": true})
	if err != nil {
		t.Errorf("nil schema should pass: %v", err)
	}
}

func TestValidateAgainstSchemaEmptySchema(t *testing.T) {
	err := ValidateAgainstSchema(json.RawMessage(`{}`), map[string]any{"anything": true})
	if err != nil {
		t.Errorf("empty schema should pass: %v", err)
	}
}

func TestValidateAgainstSchemaInvalidSchemaJSON(t *testing.T) {
	err := ValidateAgainstSchema(json.RawMessage(`not json`), map[string]any{"x": 1})
	if err != nil {
		t.Errorf("invalid schema JSON should be silently ignored: %v", err)
	}
}

func TestMapKeysViaShared(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := shared.MapKeys(m)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}
