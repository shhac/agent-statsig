package project

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/config"
	"github.com/shhac/agent-statsig/internal/credential"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/agent-statsig/internal/output"
)

func Register(root *cobra.Command) {
	proj := &cobra.Command{
		Use:   "project",
		Short: "Manage Statsig project connections",
	}

	registerAdd(proj)
	registerUpdate(proj)
	registerRemove(proj)
	registerList(proj)
	registerSetDefault(proj)
	registerTest(proj)

	root.AddCommand(proj)
}

func registerAdd(parent *cobra.Command) {
	var consoleKey, clientKey string

	cmd := &cobra.Command{
		Use:   "add <alias>",
		Short: "Add a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			if consoleKey == "" {
				output.WriteError(os.Stderr, agenterrors.New("--console-key is required", agenterrors.FixableByAgent))
				return nil
			}

			cred := credential.Credential{
				ConsoleKey: consoleKey,
				ClientKey:  clientKey,
			}
			storage, err := credential.Store(alias, cred)
			if err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				return nil
			}

			proj := config.Project{}
			if err := config.StoreProject(alias, proj); err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				return nil
			}

			output.PrintJSON(map[string]any{
				"status":  "ok",
				"alias":   alias,
				"storage": storage,
			}, true)
			return nil
		},
	}
	cmd.Flags().StringVar(&consoleKey, "console-key", "", "Statsig Console API key")
	cmd.Flags().StringVar(&clientKey, "client-key", "", "Statsig Client API key (for evaluation)")
	parent.AddCommand(cmd)
}

func registerUpdate(parent *cobra.Command) {
	var consoleKey, clientKey string

	cmd := &cobra.Command{
		Use:   "update <alias>",
		Short: "Update project credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			existing, err := credential.Get(alias)
			if err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByAgent).
					WithHint("Use 'project add' to create a new project"))
				return nil
			}

			if consoleKey != "" {
				existing.ConsoleKey = consoleKey
			}
			if clientKey != "" {
				existing.ClientKey = clientKey
			}

			storage, err := credential.Store(alias, *existing)
			if err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				return nil
			}

			output.PrintJSON(map[string]any{
				"status":  "ok",
				"alias":   alias,
				"storage": storage,
			}, true)
			return nil
		},
	}
	cmd.Flags().StringVar(&consoleKey, "console-key", "", "Statsig Console API key")
	cmd.Flags().StringVar(&clientKey, "client-key", "", "Statsig Client API key")
	parent.AddCommand(cmd)
}

func registerRemove(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			if err := credential.Remove(alias); err != nil {
				var nf *credential.NotFoundError
				if errors.As(err, &nf) {
					output.WriteError(os.Stderr, agenterrors.Newf(agenterrors.FixableByAgent, "project %q not found", alias))
				} else {
					output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				}
				return nil
			}

			config.RemoveProject(alias)

			output.PrintJSON(map[string]any{
				"status": "ok",
				"alias":  alias,
			}, true)
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Read()
			names, _ := credential.List()

			projects := make([]map[string]any, 0, len(names))
			for _, name := range names {
				entry := map[string]any{
					"alias":     name,
					"isDefault": name == cfg.DefaultProject,
				}
				projects = append(projects, entry)
			}

			output.PrintJSON(map[string]any{
				"projects": projects,
			}, true)
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerSetDefault(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "set-default <alias>",
		Short: "Set the default project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			if _, err := credential.Get(alias); err != nil {
				output.WriteError(os.Stderr, agenterrors.Newf(agenterrors.FixableByAgent, "project %q not found", alias).
					WithHint("Use 'project list' to see available projects"))
				return nil
			}

			if err := config.SetDefault(alias); err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				return nil
			}

			output.PrintJSON(map[string]any{
				"status":  "ok",
				"default": alias,
			}, true)
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func registerTest(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "test [alias]",
		Short: "Test project connectivity",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var alias string
			if len(args) > 0 {
				alias = args[0]
			}

			resolved, err := resolveAlias(alias)
			if err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}

			cred, err := credential.Get(resolved)
			if err != nil {
				output.WriteError(os.Stderr, agenterrors.Wrap(err, agenterrors.FixableByHuman))
				return nil
			}

			client := api.NewClient(cred.ConsoleKey, cred.ClientKey)
			ctx, cancel := context.WithTimeout(context.Background(), 10_000_000_000)
			defer cancel()

			_, _, err = client.ListGates(ctx, 1, 1, nil)
			if err != nil {
				output.WriteError(os.Stderr, err)
				return nil
			}

			result := map[string]any{
				"status":     "ok",
				"project":    resolved,
				"consoleKey": true,
			}

			if cred.ClientKey != "" {
				result["clientKey"] = true
			}

			output.PrintJSON(result, true)
			return nil
		},
	}
	parent.AddCommand(cmd)
}

func resolveAlias(alias string) (string, error) {
	if alias != "" {
		return alias, nil
	}
	if env := os.Getenv("AGENT_STATSIG_PROJECT"); env != "" {
		return env, nil
	}
	cfg := config.Read()
	if cfg.DefaultProject != "" {
		return cfg.DefaultProject, nil
	}
	return "", agenterrors.New("no project specified", agenterrors.FixableByAgent).
		WithHint("Specify a project alias or set a default with 'project set-default'")
}
