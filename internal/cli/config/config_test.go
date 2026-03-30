package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func TestValidateAgainstSchema(t *testing.T) {
	schema := json.RawMessage(`{
		"properties": {
			"theme": {"type": "string"},
			"limit": {"type": "number"}
		},
		"required": ["theme"]
	}`)

	err := ValidateAgainstSchema(schema, map[string]any{"theme": "dark", "limit": 10})
	if err != nil {
		t.Errorf("valid value should pass: %v", err)
	}

	err = ValidateAgainstSchema(schema, map[string]any{"limit": 10})
	if err == nil {
		t.Error("should error on missing required field")
	}

	err = ValidateAgainstSchema(schema, map[string]any{"theme": "dark", "unknown": true})
	if err == nil {
		t.Error("should error on unknown field")
	}
}

func TestValidateAgainstSchemaNoSchema(t *testing.T) {
	err := ValidateAgainstSchema(nil, map[string]any{"anything": true})
	if err != nil {
		t.Errorf("nil schema should pass: %v", err)
	}
}

func TestValidateAgainstSchemaNonObject(t *testing.T) {
	schema := json.RawMessage(`{"properties": {"x": {}}}`)
	err := ValidateAgainstSchema(schema, "not a map")
	if err != nil {
		t.Errorf("non-object value should pass (no validation): %v", err)
	}
}

func TestMapKeysViaShared(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := shared.MapKeys(m)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
}
