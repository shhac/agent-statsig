package config

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	libcli "github.com/shhac/lib-agent-cli/cli"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
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
	shared.RegisterUsage(cfg, "config", configUsage)
	libcli.HandleUnknownCommand(cfg, "run 'agent-statsig config usage' to see the available commands")

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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}

				configs, pagination, err := client.ListConfigs(ctx, limit, page, tags)
				if err != nil {
					return err
				}

				if search != "" {
					configs = shared.FilterBySearch(configs, search,
						func(c api.DynamicConfig) string { return c.Name },
						func(c api.DynamicConfig) string { return c.Description })
				}

				shared.WritePaginatedList(shared.ToAnySlice(configs), pagination, g.Format)
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

func registerGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <name>...",
		Short: "Get dynamic config details (one or more names)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetConfig(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var description string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := shared.ValidateTags(ctx, client, tags); err != nil {
					return err
				}
				cfg, err := client.CreateConfig(ctx, args[0], description, tags)
				if err != nil {
					return err
				}
				shared.WriteResource(cfg, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Config description")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable: --tag core --tag mobile)")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteConfig(ctx, args[0]); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "deleted": args[0]}, g.Format)
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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.EnableConfig(ctx, args[0]); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "config": args[0], "isEnabled": true}, g.Format)
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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.DisableConfig(ctx, args[0]); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "config": args[0], "isEnabled": false}, g.Format)
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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.ArchiveConfig(ctx, args[0]); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "config": args[0], "archived": true}, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var tags []string

	cmd := &cobra.Command{
		Use:   "update <name> <json>",
		Short: "Update a config with raw JSON (partial update)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				update, err := shared.ParseJSONArg(args[1])
				if err != nil {
					return err
				}
				if cmd.Flags().Changed("tag") {
					if err := shared.ValidateTags(ctx, client, tags); err != nil {
						return err
					}
					update["tags"] = tags
				}
				cfg, err := client.UpdateConfig(ctx, args[0], update)
				if err != nil {
					return err
				}
				shared.WriteResource(cfg, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable, replaces existing tags)")
	parent.AddCommand(cmd)
}
