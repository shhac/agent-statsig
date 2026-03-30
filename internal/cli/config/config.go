package config

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/agent-statsig/internal/output"
)

func Register(root *cobra.Command, globals func() *shared.GlobalFlags) {
	cfg := &cobra.Command{
		Use:   "config",
		Short: "Manage dynamic configs",
	}

	registerList(cfg, globals)
	registerGet(cfg, globals)
	registerCreate(cfg, globals)
	registerDelete(cfg, globals)
	registerEnable(cfg, globals)
	registerDisable(cfg, globals)
	registerArchive(cfg, globals)
	registerUpdate(cfg, globals)
	registerRule(cfg, globals)

	root.AddCommand(cfg)
}

func registerList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var limit, page int
	var tag, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List dynamic configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}

				configs, pagination, err := client.ListConfigs(ctx, limit, page, tags)
				if err != nil {
					return err
				}

				if search != "" {
					configs = filterConfigs(configs, search)
				}

				items := make([]any, len(configs))
				for i, c := range configs {
					items[i] = c
				}
				shared.WritePaginatedList(items, pagination, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "Results per page")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag (comma-separated)")
	cmd.Flags().StringVar(&search, "search", "", "Filter by name (client-side substring match)")
	parent.AddCommand(cmd)
}

func filterConfigs(configs []api.DynamicConfig, search string) []api.DynamicConfig {
	search = strings.ToLower(search)
	var filtered []api.DynamicConfig
	for _, c := range configs {
		if strings.Contains(strings.ToLower(c.Name), search) || strings.Contains(strings.ToLower(c.Description), search) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func registerGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get dynamic config details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				cfg, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}
				output.PrintJSON(cfg, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var description string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				cfg, err := client.CreateConfig(ctx, args[0], description)
				if err != nil {
					return err
				}
				output.PrintJSON(cfg, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Config description")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteConfig(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "deleted": args[0]}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerEnable(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.EnableConfig(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "config": args[0], "isEnabled": true}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerDisable(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DisableConfig(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "config": args[0], "isEnabled": false}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerArchive(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "archive <name>",
		Short: "Archive a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.ArchiveConfig(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "config": args[0], "archived": true}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "update <name> <json>",
		Short: "Update a config with raw JSON (partial update)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				var update map[string]any
				if err := json.Unmarshal([]byte(args[1]), &update); err != nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "invalid JSON: %s", err).
						WithHint("Provide a valid JSON object")
				}
				cfg, err := client.UpdateConfig(ctx, args[0], update)
				if err != nil {
					return err
				}
				output.PrintJSON(cfg, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}
