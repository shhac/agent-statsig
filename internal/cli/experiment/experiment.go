package experiment

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
	exp := &cobra.Command{
		Use:   "experiment",
		Short: "Manage experiments",
	}

	registerList(exp, globals)
	registerGet(exp, globals)
	registerCreate(exp, globals)
	registerDelete(exp, globals)
	registerArchive(exp, globals)
	registerUpdate(exp, globals)
	registerStart(exp, globals)
	registerReset(exp, globals)
	registerAbandon(exp, globals)
	registerShip(exp, globals)
	shared.RegisterUsage(exp, "experiment", experimentUsage)

	root.AddCommand(exp)
}

func registerList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var limit, page int
	var tag, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List experiments",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}

				experiments, pagination, err := client.ListExperiments(ctx, limit, page, tags)
				if err != nil {
					return err
				}

				if search != "" {
					experiments = shared.FilterBySearch(experiments, search,
						func(e api.Experiment) string { return e.Name },
						func(e api.Experiment) string { return e.Description })
				}

				shared.WritePaginatedList(shared.ToAnySlice(experiments), pagination, g.Format)
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
		Use:   "get <name>",
		Short: "Get experiment details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				exp, err := client.GetExperiment(ctx, args[0])
				if err != nil {
					return err
				}
				output.PrintJSON(exp, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerCreate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var description, groupsJSON string
	var tags []string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := shared.ValidateTags(ctx, client, tags); err != nil {
					return err
				}
				var groups []api.Group
				if groupsJSON != "" {
					if err := json.Unmarshal([]byte(groupsJSON), &groups); err != nil {
						return agenterrors.Newf(agenterrors.FixableByAgent, "invalid groups JSON: %s", err).
							WithHint(`Expected: [{"name":"control","size":50},{"name":"test","size":50}]`)
					}
				}
				exp, err := client.CreateExperiment(ctx, args[0], description, groups, tags)
				if err != nil {
					return err
				}
				output.PrintJSON(exp, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Experiment description")
	cmd.Flags().StringVar(&groupsJSON, "groups", "", "Groups JSON array")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable: --tag core --tag mobile)")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteExperiment(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "deleted": args[0]}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerArchive(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "archive <name>",
		Short: "Archive an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.ArchiveExperiment(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "experiment": args[0], "archived": true}, true)
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
		Short: "Update an experiment with raw JSON (partial update)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
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
				exp, err := client.UpdateExperiment(ctx, args[0], update)
				if err != nil {
					return err
				}
				output.PrintJSON(exp, true)
				return nil
			})
		},
	}
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable, replaces existing tags)")
	parent.AddCommand(cmd)
}

func registerStart(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.StartExperiment(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "experiment": args[0], "started": true}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerReset(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "reset <name>",
		Short: "Reset an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.ResetExperiment(ctx, args[0]); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "experiment": args[0], "reset": true}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerAbandon(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var reason string

	cmd := &cobra.Command{
		Use:   "abandon <name>",
		Short: "Abandon an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.AbandonExperiment(ctx, args[0], reason); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "experiment": args[0], "abandoned": true}, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "Decision reason")
	cmd.MarkFlagRequired("reason")
	parent.AddCommand(cmd)
}

func registerShip(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var groupID, reason string
	var removeTargeting bool

	cmd := &cobra.Command{
		Use:   "ship <name>",
		Short: "Ship (make decision on) an experiment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.ShipExperiment(ctx, args[0], groupID, reason, removeTargeting); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{
					"status":     "ok",
					"experiment": args[0],
					"shipped":    groupID,
				}, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&groupID, "group", "", "Group ID to ship")
	cmd.MarkFlagRequired("group")
	cmd.Flags().StringVar(&reason, "reason", "", "Decision reason")
	cmd.MarkFlagRequired("reason")
	cmd.Flags().BoolVar(&removeTargeting, "remove-targeting", false, "Remove targeting after shipping")
	parent.AddCommand(cmd)
}
