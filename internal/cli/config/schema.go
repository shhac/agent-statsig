package config

import (
	"context"
	"encoding/json"

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
				shared.WriteResource(map[string]any{"name": cfg.Name, "schema": schemaObjectForm(cfg.Schema)}, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

// schemaObjectForm decodes a config's stored schema into its object form for
// display: the decoded value, or the raw string if it will not decode, or nil
// when no schema is set.
func schemaObjectForm(schema json.RawMessage) any {
	normalized, ok := NormalizeSchema(schema)
	if !ok {
		return nil
	}
	var v any
	if err := json.Unmarshal(normalized, &v); err != nil {
		return string(normalized)
	}
	return v
}

func registerSchemaSet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "set <config> <schema-json>",
		Short: "Set or replace the config's JSON Schema",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				schemaVal, compiled, err := parseSchemaArg(args[1])
				if err != nil {
					return err
				}

				if !force {
					cfg, err := client.GetConfig(ctx, args[0])
					if err != nil {
						return err
					}
					if violations := schemaViolations(compiled, cfg); len(violations) > 0 {
						return conformanceError(violations, "existing values do not conform to the new schema",
							"Fix the values first ('config rule update --return-value' / 'config update'), or re-run with --force to set the schema anyway")
					}
				}

				encoded, err := json.Marshal(schemaVal)
				if err != nil {
					return agenterrors.Wrap(err, agenterrors.FixableByAgent)
				}
				return applyConfigUpdate(ctx, client, g, args[0], map[string]any{"schema": string(encoded)}, dryRun)
			})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Set the schema even if existing values do not conform")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate server-side without persisting (API dryRun)")
	parent.AddCommand(cmd)
}

func registerSchemaClear(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "clear <config>",
		Short: "Remove the config's JSON Schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				return applyConfigUpdate(ctx, client, g, args[0], map[string]any{"schema": nil}, dryRun)
			})
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate server-side without persisting (API dryRun)")
	parent.AddCommand(cmd)
}
