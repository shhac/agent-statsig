package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func registerUsageCommand(root *cobra.Command) {
	usage := &cobra.Command{
		Use:   "usage",
		Short: "Show LLM-optimized reference card",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(referenceCard)
		},
	}
	root.AddCommand(usage)
}

const referenceCard = `agent-statsig — Statsig feature flag CLI for AI agents

GLOBAL FLAGS
  --project <alias>    Project alias (or set AGENT_STATSIG_PROJECT)
  --format json|yaml|jsonl  Output format (default: json)
  --timeout <ms>       Request timeout in milliseconds

PROJECT MANAGEMENT
  project add <alias> --console-key <key> [--client-key <key>]
  project update <alias> [--console-key <key>] [--client-key <key>]
  project remove <alias>
  project list
  project set-default <alias>
  project test [alias]

FEATURE GATES
  gate list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  gate get <name>
  gate create <name> [--description <text>]
  gate delete <name>
  gate enable <name>
  gate disable <name>
  gate archive <name>
  gate launch <name>
  gate check <name> --user <json>
  gate update <name> <json>
  gate rollout <name> --percent <0-100>

  gate rule list <gate>
  gate rule add <gate> --name <rule-name> --criteria <type> --values <csv> [--operator <op>] [--pass-percent N] [--environments <csv>]
  gate rule update <gate> --rule <rule-id> [--add-values <csv>] [--remove-values <csv>] [--pass-percent N]
  gate rule remove <gate> --rule <rule-id>
  gate criteria

DYNAMIC CONFIGS
  config list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  config get <name>
  config create <name> [--description <text>]
  config delete <name>
  config enable <name>
  config disable <name>
  config archive <name>
  config update <name> <json>

  config rule list <config>
  config rule add <config> --name <rule-name> --criteria <type> --values <csv> --return-value <json> [--operator <op>] [--pass-percent N] [--environments <csv>]
  config rule update <config> --rule <rule-id> [--add-values <csv>] [--remove-values <csv>] [--pass-percent N] [--return-value <json>]
  config rule remove <config> --rule <rule-id>

EXPERIMENTS
  experiment list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  experiment get <name>
  experiment create <name> [--description <text>] [--groups <json>]
  experiment delete <name>
  experiment archive <name>
  experiment update <name> <json>
  experiment start <name>
  experiment reset <name>
  experiment abandon <name> --reason <text>
  experiment ship <name> --group <group-id> --reason <text> [--remove-targeting]

SEGMENTS
  segment list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  segment get <name>
  segment create <name> [--description <text>] [--type <type>]
  segment delete <name>
  segment archive <name>
  segment ids get <name>
  segment ids add <name> --ids <csv>
  segment ids remove <name> --ids <csv>

ERROR CLASSIFICATION
  fixable_by: agent  — typo, wrong name, bad syntax (retry with fix)
  fixable_by: human  — missing credentials, permission denied
  fixable_by: retry  — network error, rate limit, server error
`
