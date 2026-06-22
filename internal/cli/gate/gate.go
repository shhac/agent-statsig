package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	libcli "github.com/shhac/lib-agent-cli/cli"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func Register(root *cobra.Command, globals func() *shared.GlobalFlags) {
	gate := &cobra.Command{
		Use:   "gate",
		Short: "Manage feature gates",
	}

	registerList(gate, globals)
	registerGet(gate, globals)
	registerCreate(gate, globals)
	registerDelete(gate, globals)
	registerEnable(gate, globals)
	registerDisable(gate, globals)
	registerArchive(gate, globals)
	registerLaunch(gate, globals)
	registerUpdate(gate, globals)
	registerRollout(gate, globals)
	registerCheck(gate, globals)
	registerCriteria(gate)
	registerRule(gate, globals)
	shared.RegisterUsage(gate, "gate", gateUsage)
	libcli.HandleUnknownCommand(gate, "run 'agent-statsig gate usage' to see the available commands")

	root.AddCommand(gate)
}

func registerList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var limit, page int
	var tag, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List feature gates",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				var tags []string
				if tag != "" {
					tags = strings.Split(tag, ",")
				}

				gates, pagination, err := client.ListGates(ctx, limit, page, tags)
				if err != nil {
					return err
				}

				if search != "" {
					gates = shared.FilterBySearch(gates, search,
						func(g api.Gate) string { return g.Name },
						func(g api.Gate) string { return g.Description })
				}

				shared.WritePaginatedList(shared.ToAnySlice(gates), pagination, g.Format)
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
		Short: "Get feature gate details (one or more names)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.GetEntities(globals(), args, func(ctx context.Context, client *api.Client, id string) (any, error) {
				return client.GetGate(ctx, id)
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
		Short: "Create a new feature gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := shared.ValidateTags(ctx, client, tags); err != nil {
					return err
				}
				gate, err := client.CreateGate(ctx, args[0], description, tags)
				if err != nil {
					return err
				}
				shared.WriteResource(gate, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Gate description")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable: --tag core --tag mobile)")
	parent.AddCommand(cmd)
}

func registerDelete(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "delete <name>", "Delete a feature gate", "deleted", nil,
		func(ctx context.Context, client *api.Client, id string) error {
			return client.DeleteGate(ctx, id)
		})
}

func registerEnable(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "enable <name>", "Enable a feature gate", "gate", map[string]any{"isEnabled": true},
		func(ctx context.Context, client *api.Client, id string) error {
			return client.EnableGate(ctx, id)
		})
}

func registerDisable(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "disable <name>", "Disable a feature gate", "gate", map[string]any{"isEnabled": false},
		func(ctx context.Context, client *api.Client, id string) error {
			return client.DisableGate(ctx, id)
		})
}

func registerArchive(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "archive <name>", "Archive a feature gate", "gate", map[string]any{"archived": true},
		func(ctx context.Context, client *api.Client, id string) error {
			return client.ArchiveGate(ctx, id)
		})
}

func registerLaunch(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	shared.RegisterAction(parent, globals, "launch <name>", "Launch a feature gate", "gate", map[string]any{"launched": true},
		func(ctx context.Context, client *api.Client, id string) error {
			return client.LaunchGate(ctx, id)
		})
}

func registerUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var tags []string

	cmd := &cobra.Command{
		Use:   "update <name> <json>",
		Short: "Update a gate with raw JSON (partial update)",
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
				gate, err := client.UpdateGate(ctx, args[0], update)
				if err != nil {
					return err
				}
				shared.WriteResource(gate, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "Tag to apply (repeatable, replaces existing tags)")
	parent.AddCommand(cmd)
}

func registerCheck(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var user string

	cmd := &cobra.Command{
		Use:   "check <name>",
		Short: "Evaluate gate for a user (requires client key)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if !client.HasClientKey() {
					return agenterrors.New("client key required for evaluation", agenterrors.FixableByHuman).
						WithHint("Add a client key: 'project update <alias> --client-key <key>'")
				}

				var userObj map[string]any
				if err := json.Unmarshal([]byte(user), &userObj); err != nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "invalid user JSON: %s", err)
				}

				fmt.Fprintf(os.Stderr, "Note: gate evaluation via Console API is not supported. Use the gate 'get' command to inspect rules.\n")
				shared.WriteResource(map[string]any{
					"hint": "Use 'gate get' to inspect rules and evaluate locally",
				}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&user, "user", "", "User JSON object")
	cmd.MarkFlagRequired("user")
	parent.AddCommand(cmd)
}
