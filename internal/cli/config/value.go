package config

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func registerValue(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	valueCmd := &cobra.Command{
		Use:   "value",
		Short: "Manage a dynamic config's defaultValue (the fallback when no rule matches)",
	}

	registerValueGet(valueCmd, globals)
	registerValueSet(valueCmd, globals)

	parent.AddCommand(valueCmd)
}

func registerValueGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <config>",
		Short: "Show the config's defaultValue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				cfg, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"name": cfg.Name, "defaultValue": decodeJSONValue(cfg.DefaultValue)}, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerValueSet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var force bool

	cmd := &cobra.Command{
		Use:   "set <config> <json>",
		Short: "Set the config's defaultValue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				var value any
				if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "invalid defaultValue JSON: %s", err).
						WithHint("Provide a JSON object, e.g. '{\"theme\":\"light\"}'")
				}
				if _, ok := value.(map[string]any); !ok {
					return agenterrors.New("defaultValue must be a JSON object", agenterrors.FixableByAgent).
						WithHint("Dynamic configs return JSON objects; wrap scalars in a field, e.g. '{\"value\":42}'")
				}

				if !force {
					cfg, err := client.GetConfig(ctx, args[0])
					if err != nil {
						return err
					}
					if err := ValidateAgainstSchema(cfg.Schema, value); err != nil {
						return err
					}
				}

				updated, err := client.UpdateConfig(ctx, args[0], map[string]any{"defaultValue": value})
				if err != nil {
					return err
				}
				shared.WriteResource(updated, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Skip client-side schema validation")
	parent.AddCommand(cmd)
}
