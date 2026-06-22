package segment

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	libcli "github.com/shhac/lib-agent-cli/cli"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func Register(root *cobra.Command, globals func() *shared.GlobalFlags) {
	seg := &cobra.Command{
		Use:   "segment",
		Short: "Manage segments",
	}

	registerList(seg, globals)
	registerGet(seg, globals)
	registerCreate(seg, globals)
	registerDelete(seg, globals)
	registerArchive(seg, globals)
	registerIDs(seg, globals)
	shared.RegisterUsage(seg, "segment", segmentUsage)
	libcli.HandleUnknownCommand(seg, "run 'agent-statsig segment usage' to see the available commands")

	root.AddCommand(seg)
}

func registerList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var limit, page int
	var tag, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List segments",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}

				segments, pagination, err := client.ListSegments(ctx, limit, page, tags)
				if err != nil {
					return err
				}

				if search != "" {
					segments = shared.FilterBySearch(segments, search,
						func(s api.Segment) string { return s.Name },
						func(s api.Segment) string { return s.Description })
				}

				shared.WritePaginatedList(shared.ToAnySlice(segments), pagination, g.Format)
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
		Short: "Get segment details (one or more names)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetSegment(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var description, segType string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				seg, err := client.CreateSegment(ctx, args[0], description, segType)
				if err != nil {
					return err
				}
				shared.WriteResource(seg, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Segment description")
	cmd.Flags().StringVar(&segType, "type", "", "Segment type: id_list, rule_based")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "delete <name>", "Delete a segment", "deleted", nil,
		func(ctx context.Context, client *api.Client, id string) error {
			return client.DeleteSegment(ctx, id)
		})
}

func registerArchive(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "archive <name>", "Archive a segment", "segment", map[string]any{"archived": true},
		func(ctx context.Context, client *api.Client, id string) error {
			return client.ArchiveSegment(ctx, id)
		})
}

func registerIDs(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	ids := &cobra.Command{
		Use:   "ids",
		Short: "Manage segment ID lists",
	}

	registerIDsGet(ids, globals)
	registerIDsAdd(ids, globals)
	registerIDsRemove(ids, globals)

	parent.AddCommand(ids)
}

func registerIDsGet(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "get <segment>...",
		Short: "Get IDs in a segment (one or more segment names)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetSegmentIDs(ctx, id)
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerIDsAdd(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ids []string

	cmd := &cobra.Command{
		Use:   "add <segment>",
		Short: "Add IDs to a segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.AddSegmentIDs(ctx, args[0], ids); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "segment": args[0], "added": len(ids)}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringArrayVar(&ids, "id", nil, "ID to add (repeatable: --id user1 --id user2)")
	cmd.MarkFlagRequired("id")
	parent.AddCommand(cmd)
}

func registerIDsRemove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ids []string

	cmd := &cobra.Command{
		Use:   "remove <segment>",
		Short: "Remove IDs from a segment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := client.RemoveSegmentIDs(ctx, args[0], ids); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "segment": args[0], "removed": len(ids)}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringArrayVar(&ids, "id", nil, "ID to remove (repeatable: --id user1 --id user2)")
	cmd.MarkFlagRequired("id")
	parent.AddCommand(cmd)
}
