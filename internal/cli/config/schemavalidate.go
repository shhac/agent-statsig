package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/shhac/agent-statsig/internal/api"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/agent-statsig/internal/output"
)

// This file is the config package's schema-validation library: everything
// pure (or client-reading) that enforces a config's JSON Schema. The cobra
// commands that expose it live in schema.go.

// ValidateAgainstSchema validates a value against a JSON Schema using full
// spec compliance. Accepts both the API's string-form schema and object form
// (see NormalizeSchema); a missing or malformed schema skips validation.
func ValidateAgainstSchema(schema json.RawMessage, value any) error {
	compiled, err := compileStoredSchema(schema, false)
	if err != nil || compiled == nil {
		return err
	}

	if err := compiled.Validate(value); err != nil {
		return agenterrors.Newf(agenterrors.FixableByAgent, "return value does not match config schema: %s", err).
			WithHint("Check the config's schema with 'config get <name>'")
	}
	return nil
}

// NormalizeSchema returns the object form of a config schema. The Console API
// stores schemas string-form (JSON-encoded text), while older fixtures and
// dryRun payloads may carry object form — accept both. The bool reports
// whether a usable schema is present.
func NormalizeSchema(schema json.RawMessage) (json.RawMessage, bool) {
	trimmed := bytes.TrimSpace(schema)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, false
	}
	if trimmed[0] != '"' {
		return trimmed, true
	}
	var inner string
	if err := json.Unmarshal(trimmed, &inner); err != nil {
		return nil, false
	}
	inner = strings.TrimSpace(inner)
	if inner == "" || inner == "null" {
		return nil, false
	}
	return json.RawMessage(inner), true
}

// draft202012URI is the exact $schema value Statsig requires on a stored
// schema — verified against the live API, which exact-matches this string and
// rejects any other. The CLI supplies it so callers never have to.
//
// parseSchemaArg still rejects other drafts client-side rather than deferring
// to that server check: the local compile happens before the request, so a
// declared draft-07 would compile with draft-07 semantics and could fail the
// existing-values pre-flight with a misleading error first.
const draft202012URI = "https://json-schema.org/draft/2020-12/schema"

// parseSchemaArg parses a user-supplied schema argument: valid JSON, object
// form, draft 2020-12 only, and compilable. Returns the parsed value (for
// re-encoding into the API's string form) alongside the compiled schema.
func parseSchemaArg(raw string) (any, *jsonschema.Schema, error) {
	var schemaVal any
	if err := json.Unmarshal([]byte(raw), &schemaVal); err != nil {
		return nil, nil, agenterrors.Newf(agenterrors.FixableByAgent, "invalid schema JSON: %s", err).
			WithHint("Provide the schema as a JSON object, e.g. '{\"type\":\"object\",\"required\":[\"theme\"]}'")
	}
	schemaMap, ok := schemaVal.(map[string]any)
	if !ok {
		return nil, nil, agenterrors.New("schema must be a JSON object", agenterrors.FixableByAgent).
			WithHint("Pass an object-form JSON Schema; the CLI handles the API's string encoding for you")
	}
	if declared := schemaMap["$schema"]; declared != nil {
		uri, isString := declared.(string)
		if !isString || !isDraft202012(uri) {
			return nil, nil, agenterrors.Newf(agenterrors.FixableByAgent, "unsupported $schema %s: Statsig evaluates schemas as JSON Schema draft 2020-12 only", describeJSON(declared)).
				WithHint("Remove $schema — the CLI adds the draft 2020-12 URI for you. Older drafts are not a safe subset — e.g. draft-07 tuple-form 'items' means something different in 2020-12")
		}
	}
	schemaMap["$schema"] = draft202012URI

	compiled, err := compileSchemaValue(schemaVal)
	if err != nil {
		return nil, nil, agenterrors.Newf(agenterrors.FixableByAgent, "not a valid JSON Schema (draft 2020-12): %s", err).
			WithHint("Fix the schema; see https://json-schema.org/draft/2020-12")
	}
	return schemaVal, compiled, nil
}

// describeJSON renders a rejected JSON value for an error message, quoting
// strings so a bare URI is not mistaken for prose.
func describeJSON(v any) string {
	if s, ok := v.(string); ok {
		return fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("%v (%T)", v, v)
}

// isDraft202012 reports whether a $schema URI declares JSON Schema draft
// 2020-12 — the only draft Statsig evaluates.
func isDraft202012(uri string) bool {
	uri = strings.TrimSuffix(strings.TrimSpace(uri), "#")
	return uri == draft202012URI ||
		uri == "http://json-schema.org/draft/2020-12/schema"
}

// noticeSchemaUnvalidatable warns, non-fatally on stderr, that a config's
// stored schema could not be compiled locally — an unknown JSON Schema draft
// or a malformed schema — so client-side validation was skipped for this
// write. The operation still proceeds; Statsig enforces the schema server-side.
// This keeps "unreadable schema" from silently masquerading as "no schema".
func noticeSchemaUnvalidatable(err error) {
	output.WriteNotice(os.Stderr,
		fmt.Sprintf("config's stored schema could not be validated locally (%s); skipping client-side validation — Statsig still enforces it server-side", err),
		"If the schema uses a newer JSON Schema draft, update agent-statsig; otherwise inspect it with 'config schema get <name>'")
}

// compileStoredSchema resolves a raw config schema into a compiled validator
// for client-side checks, encoding the single policy both validation entry
// points share. It returns (nil, nil) — "nothing to validate against" — when
// the schema is absent, or when it cannot be compiled locally and userSupplied
// is false (an unreadable *stored* schema, e.g. a newer draft): a non-fatal
// notice is emitted and the server arbitrates. A compile failure on a
// user-supplied schema is a hard error instead.
func compileStoredSchema(schema json.RawMessage, userSupplied bool) (*jsonschema.Schema, error) {
	normalized, ok := NormalizeSchema(schema)
	if !ok {
		return nil, nil
	}
	compiled, err := compileSchema(normalized)
	if err != nil {
		if userSupplied {
			return nil, agenterrors.Newf(agenterrors.FixableByAgent, "not a valid JSON Schema (draft 2020-12): %s", err).
				WithHint("Fix the schema, or use 'config schema set <name> <json>'")
		}
		noticeSchemaUnvalidatable(err)
		return nil, nil
	}
	return compiled, nil
}

func compileSchema(raw json.RawMessage) (*jsonschema.Schema, error) {
	var schemaVal any
	if err := json.Unmarshal(raw, &schemaVal); err != nil {
		return nil, err
	}
	return compileSchemaValue(schemaVal)
}

func compileSchemaValue(schemaVal any) (*jsonschema.Schema, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", schemaVal); err != nil {
		return nil, err
	}
	return compiler.Compile("schema.json")
}

// schemaViolations validates a config's stored defaultValue and rule return
// values against a schema. Absent/null values are skipped — there is nothing
// to validate — but present empty objects are validated like any other value.
func schemaViolations(compiled *jsonschema.Schema, cfg *api.DynamicConfig) []string {
	var violations []string
	if dv := decodeJSONValue(cfg.DefaultValue); dv != nil {
		if err := compiled.Validate(dv); err != nil {
			violations = append(violations, fmt.Sprintf("defaultValue: %v", err))
		}
	}
	for _, r := range cfg.Rules {
		if r.ReturnValue == nil {
			continue
		}
		label := r.Name
		if r.ID != "" {
			label = fmt.Sprintf("%s (%s)", r.Name, r.ID)
		}
		if err := compiled.Validate(r.ReturnValue); err != nil {
			violations = append(violations, fmt.Sprintf("rule %q returnValue: %v", label, err))
		}
	}
	return violations
}

func decodeJSONValue(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}

// validateUpdatePayload runs client-side schema checks on a raw 'config
// update' body: a schema field must be string-form (the API rejects objects),
// and any defaultValue / rule returnValues in the payload must conform to the
// effective schema — the one being set, or else the stored one.
func validateUpdatePayload(ctx context.Context, client *api.Client, id string, update map[string]any, force bool) error {
	if err := validateSchemaFieldShape(update); err != nil {
		return err
	}
	if force {
		return nil
	}

	schemaField, hasSchema := update["schema"]
	_, hasDV := update["defaultValue"]
	_, hasRules := update["rules"]
	if !hasDV && !hasRules && !hasSchema {
		return nil
	}

	schemaRaw, err := effectiveSchema(ctx, client, id, schemaField, hasSchema)
	if err != nil || schemaRaw == nil {
		return err
	}
	compiled, err := compileStoredSchema(schemaRaw, hasSchema)
	if err != nil || compiled == nil {
		return err
	}

	partial, ok := decodePartialConfig(update)
	if !ok {
		return nil // unrecognizable payload shapes skip client-side validation; the server arbitrates
	}
	if violations := schemaViolations(compiled, partial); len(violations) > 0 {
		return conformanceError(violations, "update does not conform to the config's schema",
			"Fix the values, change the schema with 'config schema set', or re-run with --force to skip client-side validation")
	}
	return nil
}

// validateSchemaFieldShape guards the raw 'schema' field of a config update:
// the API stores schemas string-form, so an object is rejected before the
// request. Runs regardless of --force, since it catches a client bug rather
// than a value-conformance issue.
func validateSchemaFieldShape(update map[string]any) error {
	schemaField, hasSchema := update["schema"]
	if !hasSchema {
		return nil
	}
	switch schemaField.(type) {
	case string, nil:
		return nil
	default:
		return agenterrors.New("schema in a raw update must be a JSON-encoded string", agenterrors.FixableByAgent).
			WithHint("Prefer 'config schema set <name> <json>' — it handles the API's string encoding and validates existing values")
	}
}

// conformanceError builds the standard "values don't match the schema" error
// from a non-empty violations list, letting each caller frame it with its own
// message and hint.
func conformanceError(violations []string, msg, hint string) error {
	return agenterrors.Newf(agenterrors.FixableByAgent, "%s: %s", msg, strings.Join(violations, "; ")).
		WithHint(hint)
}

// effectiveSchema resolves which schema a raw update should be validated
// against: the one being set, else the stored one. A nil result (with nil
// error) means validation is off — the update clears the schema.
func effectiveSchema(ctx context.Context, client *api.Client, id string, schemaField any, hasSchema bool) (json.RawMessage, error) {
	if hasSchema {
		s, ok := schemaField.(string)
		if !ok {
			return nil, nil // schema: null clears
		}
		return json.RawMessage(s), nil
	}
	cfg, err := client.GetConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	return cfg.Schema, nil
}

// decodePartialConfig round-trips an update's defaultValue/rules into the
// typed shape so schemaViolations stays the single walker for both stored
// configs and raw update payloads.
func decodePartialConfig(update map[string]any) (*api.DynamicConfig, bool) {
	subset := map[string]any{}
	if dv, ok := update["defaultValue"]; ok && dv != nil {
		subset["defaultValue"] = dv
	}
	if rules, ok := update["rules"]; ok && rules != nil {
		subset["rules"] = rules
	}
	b, err := json.Marshal(subset)
	if err != nil {
		return nil, false
	}
	var partial api.DynamicConfig
	if err := json.Unmarshal(b, &partial); err != nil {
		return nil, false
	}
	return &partial, true
}
