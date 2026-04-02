package tag

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	"github.com/shhac/agent-statsig/internal/output"
)

func Register(root *cobra.Command, globals func() *shared.GlobalFlags) {
	tag := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}

	registerList(tag, globals)
	registerGet(tag, globals)
	registerCreate(tag, globals)
	registerUpdate(tag, globals)
	registerDelete(tag, globals)
	registerUsage(tag, globals)

	root.AddCommand(tag)
}

func registerList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var limit, page int
	var search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				tags, pagination, err := client.ListTags(ctx, limit, page)
				if err != nil {
					return err
				}

				if search != "" {
					tags = shared.FilterBySearch(tags, search,
						func(t api.Tag) string { return t.Name },
						func(t api.Tag) string { return t.Description })
				}

				shared.WritePaginatedList(shared.ToAnySlice(tags), pagination, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "Results per page")
	cmd.Flags().IntVar(&page, "page", 0, "Page number")
	cmd.Flags().StringVar(&search, "search", "", "Filter by name (client-side substring match)")
	parent.AddCommand(cmd)
}

func registerGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get tag details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				tag, err := client.GetTag(ctx, args[0])
				if err != nil {
					return err
				}
				output.PrintJSON(tag, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var description string
	var isCore bool

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				tag, err := client.CreateTag(ctx, args[0], description, isCore)
				if err != nil {
					return err
				}
				output.PrintJSON(tag, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Tag description")
	cmd.Flags().BoolVar(&isCore, "is-core", false, "Mark as a core tag")
	parent.AddCommand(cmd)
}

func registerUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var name, description string
	var isCore bool

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				update := map[string]any{}
				if cmd.Flags().Changed("name") {
					update["name"] = name
				}
				if cmd.Flags().Changed("description") {
					update["description"] = description
				}
				if cmd.Flags().Changed("is-core") {
					update["isCore"] = isCore
				}
				tag, err := client.UpdateTag(ctx, args[0], update)
				if err != nil {
					return err
				}
				output.PrintJSON(tag, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Tag name")
	cmd.Flags().StringVar(&description, "description", "", "Tag description")
	cmd.Flags().BoolVar(&isCore, "is-core", false, "Mark as a core tag")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteTag(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "deleted": args[0]}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}
