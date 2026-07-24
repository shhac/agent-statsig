package shared

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/shhac/agent-statsig/internal/api"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

// ValidateTags checks that all specified tag names exist in the project.
// Returns a helpful error before any mutation if tags are missing.
func ValidateTags(ctx context.Context, client *api.Client, tagNames []string) error {
	if len(tagNames) == 0 {
		return nil
	}
	tags, _, err := client.ListTags(ctx, 0, 0)
	if err != nil {
		return err
	}
	existing := make(map[string]bool, len(tags))
	for _, t := range tags {
		existing[t.Name] = true
	}
	var missing []string
	for _, name := range tagNames {
		if !existing[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return agenterrors.Newf(agenterrors.FixableByAgent,
			"tags not found: %s", strings.Join(missing, ", ")).
			WithHint("Create missing tags first with 'tag create <name>', or use 'tag list' to see available tags")
	}
	return nil
}

// ValidateCriteria checks that a criteria type and operator are valid Statsig condition values.
func ValidateCriteria(criteria, operator string) error {
	found := false
	for _, ct := range api.ConditionTypes {
		if ct == criteria {
			found = true
			break
		}
	}
	if !found {
		return agenterrors.Newf(agenterrors.FixableByAgent, "unknown criteria %q", criteria).
			WithHint("Use 'gate criteria' to list available criteria types")
	}

	if operator == "" {
		return nil
	}

	ops, ok := api.OperatorsByType[criteria]
	if !ok || len(ops) == 0 {
		return nil
	}

	for _, op := range ops {
		if op == operator {
			return nil
		}
	}
	return agenterrors.Newf(agenterrors.FixableByAgent, "invalid operator %q for criteria %q", operator, criteria).
		WithHint("Valid operators: " + strings.Join(ops, ", "))
}

// Slice helpers

// ParseJSONArg parses a JSON string argument into a map, returning a classified error on failure.
func ParseJSONArg(raw string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, agenterrors.Newf(agenterrors.FixableByAgent, "invalid JSON: %s", err).
			WithHint("Provide a valid JSON object")
	}
	return result, nil
}

// ParseJSONValue parses a JSON string argument into an arbitrary value,
// returning a classified error that names the argument on failure. label is the
// argument's name in the message (e.g. "return-value", "defaultValue",
// "schema"). Use ParseJSONArg instead when the value must be a JSON object.
func ParseJSONValue(raw, label string) (any, error) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, agenterrors.Newf(agenterrors.FixableByAgent, "invalid %s JSON: %s", label, err).
			WithHint("Provide valid JSON")
	}
	return v, nil
}
