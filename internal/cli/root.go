package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cliconfig "github.com/shhac/agent-statsig/internal/cli/config"
	"github.com/shhac/agent-statsig/internal/cli/experiment"
	"github.com/shhac/agent-statsig/internal/cli/gate"
	"github.com/shhac/agent-statsig/internal/cli/project"
	"github.com/shhac/agent-statsig/internal/cli/segment"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	"github.com/shhac/agent-statsig/internal/cli/tag"
)

var (
	flagProject string
	flagFormat  string
	flagTimeout int
)

func allGlobals() *shared.GlobalFlags {
	return &shared.GlobalFlags{
		Project: flagProject,
		Format:  flagFormat,
		Timeout: flagTimeout,
	}
}

func newRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "agent-statsig",
		Short:         "Statsig feature flag CLI for AI agents",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "Project alias (or set AGENT_STATSIG_PROJECT)")
	root.PersistentFlags().StringVarP(&flagFormat, "format", "f", "", "Output format: json, yaml, jsonl")
	root.PersistentFlags().IntVarP(&flagTimeout, "timeout", "t", 0, "Request timeout in milliseconds")

	registerUsageCommand(root)
	project.Register(root)
	gate.Register(root, allGlobals)
	cliconfig.Register(root, allGlobals)
	experiment.Register(root, allGlobals)
	segment.Register(root, allGlobals)
	tag.Register(root, allGlobals)

	return root
}

func Execute(version string) error {
	err := newRootCmd(version).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
