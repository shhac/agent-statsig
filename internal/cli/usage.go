package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func registerUsageCommand(root *cobra.Command) {
	usage := &cobra.Command{
		Use:   "llm-help",
		Short: "Show LLM-optimized reference card",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(referenceCard)
		},
	}
	root.AddCommand(usage)
}

const referenceCard = `agent-statsig — Statsig feature flag CLI for AI agents

All output is structured JSON to stdout. Errors are JSON to stderr with
fixable_by classification (agent/human/retry). Use --format yaml or
--format jsonl for alternative output.

QUICK START (read-only — safe to explore)
  gate list                          List all feature gates
  gate get <name>                    Inspect a gate's rules and state
  config list                        List all dynamic configs
  config get <name>                  Inspect a config's rules and values
  experiment list                    List all experiments
  segment list                       List all segments

COMMON WORKFLOWS
  To roll out a gate to a percentage:
    gate rollout <name> --percent 50

  To target specific users:
    gate rule add <name> --name "Team" --criteria email --value user@co.com

  To inspect before modifying:
    gate get <name>                  ← read rules first
    gate rule list <name>            ← see rule IDs
    gate rule update <name> --rule <id> --add-value new@co.com

  To modify a dynamic config's value:
    config get <name>                ← check schema + current rules
    config rule add <name> --name "Rule" --criteria email --value user@co.com --return-value '{"key":"val"}'

GLOBAL FLAGS
  -p, --project <alias>              Project alias (or AGENT_STATSIG_PROJECT env)
  --format json|yaml|jsonl           Output format (default: json)
  --timeout <ms>                     Request timeout in milliseconds

PER-ENTITY REFERENCE (run these for detailed help + examples)
  gate llm-help                   Feature gates reference
  config llm-help                 Dynamic configs reference
  experiment llm-help             Experiments reference
  segment llm-help                Segments reference

PROJECT MANAGEMENT
  project add <alias> --console-key <key> [--client-key <key>]
  project update <alias> [--console-key <key>] [--client-key <key>]
  project remove <alias>
  project list
  project set-default <alias>
  project test [alias]

ERROR HANDLING
  Errors include a hint and classification:
    fixable_by: agent  → typo, wrong name, bad syntax — retry with fix
    fixable_by: human  → missing credentials, permission denied
    fixable_by: retry  → network error, rate limit, server error

RULE CONCEPTS
  Rules are evaluated top-to-bottom; first matching rule wins.
  Conditions within a rule are AND-ed together.
  Use 'gate criteria' to list all 25 condition types and their operators.
  Default operator is 'any' (case-insensitive match).
`
