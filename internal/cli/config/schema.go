package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func registerSchema(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Manage a dynamic config's JSON Schema (draft 2020-12)",
	}

	registerSchemaGet(schemaCmd, globals)
	registerSchemaSet(schemaCmd, globals)
	registerSchemaClear(schemaCmd, globals)

	parent.AddCommand(schemaCmd)
}

func registerSchemaGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <config>",
		Short: "Show the config's JSON Schema in object form",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				cfg, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}
				var schemaVal any
				if normalized, ok := NormalizeSchema(cfg.Schema); ok {
					if err := json.Unmarshal(normalized, &schemaVal); err != nil {
						schemaVal = string(normalized)
					}
				}
				shared.WriteResource(map[string]any{"name": cfg.Name, "schema": schemaVal}, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerSchemaSet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var force bool

	cmd := &cobra.Command{
		Use:   "set <config> <schema-json>",
		Short: "Set or replace the config's JSON Schema",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				var schemaVal any
				if err := json.Unmarshal([]byte(args[1]), &schemaVal); err != nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "invalid schema JSON: %s", err).
						WithHint("Provide the schema as a JSON object, e.g. '{\"type\":\"object\",\"required\":[\"theme\"]}'")
				}
				if _, ok := schemaVal.(map[string]any); !ok {
					return agenterrors.New("schema must be a JSON object", agenterrors.FixableByAgent).
						WithHint("Pass an object-form JSON Schema; the CLI handles the API's string encoding for you")
				}
				compiled, err := compileSchemaValue(schemaVal)
				if err != nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "not a valid JSON Schema (draft 2020-12): %s", err).
						WithHint("Fix the schema; see https://json-schema.org/draft/2020-12")
				}

				if !force {
					cfg, err := client.GetConfig(ctx, args[0])
					if err != nil {
						return err
					}
					if violations := schemaViolations(compiled, cfg); len(violations) > 0 {
						return agenterrors.Newf(agenterrors.FixableByAgent,
							"existing values do not conform to the new schema: %s", strings.Join(violations, "; ")).
							WithHint("Fix the values first ('config rule update --return-value' / 'config update'), or re-run with --force to set the schema anyway")
					}
				}

				encoded, err := json.Marshal(schemaVal)
				if err != nil {
					return agenterrors.Wrap(err, agenterrors.FixableByAgent)
				}
				updated, err := client.UpdateConfig(ctx, args[0], map[string]any{"schema": string(encoded)})
				if err != nil {
					return err
				}
				shared.WriteResource(updated, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Set the schema even if existing values do not conform")
	parent.AddCommand(cmd)
}

func registerSchemaClear(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "clear <config>",
		Short: "Remove the config's JSON Schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				updated, err := client.UpdateConfig(ctx, args[0], map[string]any{"schema": nil})
				if err != nil {
					return err
				}
				shared.WriteResource(updated, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
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
	schemaField, hasSchema := update["schema"]
	if hasSchema {
		switch schemaField.(type) {
		case string, nil:
		default:
			return agenterrors.New("schema in a raw update must be a JSON-encoded string", agenterrors.FixableByAgent).
				WithHint("Prefer 'config schema set <name> <json>' — it handles the API's string encoding and validates existing values")
		}
	}
	if force {
		return nil
	}

	dv, hasDV := update["defaultValue"]
	rules, hasRules := update["rules"]
	if !hasDV && !hasRules && !hasSchema {
		return nil
	}

	var schemaRaw json.RawMessage
	if hasSchema {
		s, ok := schemaField.(string)
		if !ok {
			return nil // schema: null clears — nothing to validate against
		}
		schemaRaw = json.RawMessage(s)
	} else {
		cfg, err := client.GetConfig(ctx, id)
		if err != nil {
			return err
		}
		schemaRaw = cfg.Schema
	}

	normalized, ok := NormalizeSchema(schemaRaw)
	if !ok {
		return nil
	}
	compiled, err := compileSchema(normalized)
	if err != nil {
		if hasSchema {
			return agenterrors.Newf(agenterrors.FixableByAgent, "not a valid JSON Schema (draft 2020-12): %s", err).
				WithHint("Fix the schema, or use 'config schema set <name> <json>'")
		}
		return nil
	}

	var violations []string
	if hasDV && dv != nil {
		if err := compiled.Validate(dv); err != nil {
			violations = append(violations, fmt.Sprintf("defaultValue: %v", err))
		}
	}
	if hasRules {
		if ruleList, ok := rules.([]any); ok {
			for _, r := range ruleList {
				rule, ok := r.(map[string]any)
				if !ok {
					continue
				}
				rv, ok := rule["returnValue"]
				if !ok || rv == nil {
					continue
				}
				name, _ := rule["name"].(string)
				if err := compiled.Validate(rv); err != nil {
					violations = append(violations, fmt.Sprintf("rule %q returnValue: %v", name, err))
				}
			}
		}
	}
	if len(violations) > 0 {
		return agenterrors.Newf(agenterrors.FixableByAgent,
			"update does not conform to the config's schema: %s", strings.Join(violations, "; ")).
			WithHint("Fix the values, change the schema with 'config schema set', or re-run with --force to skip client-side validation")
	}
	return nil
}
