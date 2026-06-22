package cli

import (
	"github.com/spf13/cobra"

	libcli "github.com/shhac/lib-agent-cli/cli"
	agentmcp "github.com/shhac/lib-agent-mcp"

	cliconfig "github.com/shhac/agent-statsig/internal/cli/config"
	"github.com/shhac/agent-statsig/internal/cli/experiment"
	"github.com/shhac/agent-statsig/internal/cli/gate"
	"github.com/shhac/agent-statsig/internal/cli/project"
	"github.com/shhac/agent-statsig/internal/cli/segment"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	"github.com/shhac/agent-statsig/internal/cli/tag"
	"github.com/shhac/agent-statsig/internal/output"
)

func newRootCmd(version string) *cobra.Command {
	g := &shared.GlobalFlags{}

	root := libcli.NewRoot(libcli.Options{
		Use:           "agent-statsig",
		Short:         "Statsig feature flag CLI for AI agents",
		Version:       version,
		Globals:       &g.Globals,
		DefaultFormat: output.FormatJSON,
		UnknownHint:   "run 'agent-statsig usage' to see the available commands",
	})

	root.PersistentFlags().StringVarP(&g.Project, "project", "p", "", "Project alias (or set AGENT_STATSIG_PROJECT)")

	globals := func() *shared.GlobalFlags { return g }

	registerUsageCommand(root)
	project.Register(root)
	gate.Register(root, globals)
	cliconfig.Register(root, globals)
	experiment.Register(root, globals)
	segment.Register(root, globals)
	tag.Register(root, globals)

	// Expose the whole command tree as an MCP server (added last, so it reflects
	// the complete tree). --color/--expose are output-shaping, irrelevant to a
	// tool call, so hide them from the generated schemas.
	// Opt the agent-facing groups into the MCP tool surface: each becomes one
	// coarse tool that dispatches its subcommands (with a "help" verb), so the
	// surface is ~one-tool-per-group instead of one-per-leaf. Credential/config/
	// usage commands are deliberately left out — they aren't agent tasks.
	exposeGroups(root,
		"experiment", "gate", "project", "segment", "tag")

	root.AddCommand(agentmcp.Command(root, agentmcp.WithHiddenFlags("color", "expose")))

	return root
}

// Execute builds the root command and runs it via libcli.Run, which renders any
// bubbled error in the structured contract on stderr and exits 1. Command RunE
// bodies return their errors unrendered (single-sink), so every failure — a
// command-level error, an unknown subcommand, or a bad flag — is rendered
// exactly once by Run and exits non-zero.
func Execute(version string) {
	libcli.Run(newRootCmd(version))
}

// exposeGroups opts the named top-level commands into the MCP tool surface.
// A name with no matching command is skipped silently — the list is a curation
// of agent-facing groups, not a registration check.
func exposeGroups(root *cobra.Command, names ...string) {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	for _, c := range root.Commands() {
		if want[c.Name()] {
			agentmcp.Expose(c)
		}
	}
}
