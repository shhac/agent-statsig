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

All output is structured JSON to stdout. Errors are JSON to stderr with
fixable_by classification (agent/human/retry). Use --format yaml or
--format jsonl for alternative output.

QUICK START (read-only — safe to explore)
  gate list                          List all feature gates
  gate get <name>...                 Inspect one or more gates' rules and state
  config list                        List all dynamic configs
  config get <name>...               Inspect one or more configs' rules and values
  experiment list                    List all experiments
  segment list                       List all segments
  tag list                           List all tags

COMMON WORKFLOWS
  To roll out a gate to a percentage:
    gate rollout <name> --percent 50

  To target specific users:
    gate rule add <name> --name "Team" --criteria email --value user@co.com

  To inspect before modifying:
    gate get <name>...               ← read rules first (one or more gates)
    gate rule list <name>            ← see rule IDs
    gate rule update <name> --rule <id> --add-value new@co.com

  To modify a dynamic config's value:
    config get <name>...             ← check schema + current rules
    config rule add <name> --name "Rule" --criteria email --value user@co.com --return-value '{"key":"val"}'

  To enforce a JSON Schema on a config's values:
    config schema get <name>         ← current schema (object form)
    config schema set <name> '{"type":"object","required":["key"]}'
    (details + conformance policy: config usage)

  To tag entities for organization:
    tag create "mobile" --description "Mobile features" --is-core
    gate create my_gate --tag mobile

GLOBAL FLAGS
  -p, --project <alias>              Project alias (or AGENT_STATSIG_PROJECT env)
  -f, --format json|yaml|jsonl       Output format (default: json)
  -t, --timeout <ms>                 Request timeout in milliseconds
  -d, --debug                        Log [debug] METHOD URL to stderr for every API call

PER-ENTITY REFERENCE (run these for detailed help + examples)
  gate usage                      Feature gates reference
  config usage                    Dynamic configs reference
  experiment usage                Experiments reference
  segment usage                   Segments reference
  tag usage                       Tags reference

PROJECT MANAGEMENT
  project add <alias> [--console-key <key>] [--client-key <key>] [--form]
    --form prompts for missing keys via a native OS dialog (preferred: the
    agent never sees the typed secret — it stays off argv and out of context).
  project update <alias> [--console-key <key>] [--client-key <key>] [--form]
  project remove <alias>
  project list
  project set-default <alias>
  project test [alias]

OUTPUT
  List commands default to JSON. Get commands default to NDJSON (one record per id,
  in input order). Use --format json|yaml for a {"data":[…],"@unresolved":[…]} envelope.

  Get (single + multi). get <id>... takes one or more ids and returns one result per
  id, in input order. Default output is NDJSON: one line per id — the record, or
  {"@unresolved":{"id","reason","fixable_by","hint"?}} for an id that couldn't be
  resolved (e.g. not found / bad id). --format json|yaml collapses to one
  {"data":[…],"@unresolved":[…]} envelope. A single get <id> is just the one-element
  case (NDJSON one line by default; pass --format json for the object). Item-level
  misses stay on stdout and exit 0; only a command-level failure (auth, network) goes
  to stderr with exit 1 and empty stdout.

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
