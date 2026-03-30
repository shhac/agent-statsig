package gate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func registerUsage(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "llm-help",
		Short: "Show feature gates detailed reference",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(gateUsage)
		},
	}
	parent.AddCommand(cmd)
}

const gateUsage = `FEATURE GATES — Reference

A feature gate is a boolean flag that controls whether a feature is on or off.
Gates have rules that target users by criteria (email, user_id, country, etc.).
Rules are evaluated top-to-bottom; first match wins. Default is fail (off).

READ
  gate list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  gate get <name>                    Full gate with rules, conditions, rollout %
  gate rule list <gate>              List rules with their IDs
  gate criteria                      List all 25 condition types + valid operators

MODIFY
  gate create <name> [--description <text>]
  gate delete <name>
  gate enable <name>                 Turn gate on (rules start evaluating)
  gate disable <name>                Turn gate off (all users fail)
  gate archive <name>
  gate launch <name>                 Mark as launched (permanent)
  gate update <name> <json>          Raw JSON partial update (escape hatch)

ROLLOUT
  gate rollout <name> --percent <0-100> [--env staging --env production]
    Creates or updates an "Everyone" (public) rule at the given percentage.
    Example: gate rollout my_gate --percent 50

RULE MANAGEMENT
  gate rule add <gate>
    --name <rule-name>               Rule display name
    --criteria <type>                Condition type (e.g. email, user_id)
    [--operator <op>]                Default: any (case-insensitive match)
    [--value <v>]                    Target value (repeatable: --value a --value b)
    [--pass-percent N]               Default: 100
    [--env <environment>]            Repeatable: --env staging --env production
    [--field <name>]                 For custom_field criteria only

  gate rule update <gate>
    --rule <rule-id>                 Use 'gate rule list' to find IDs
    [--add-value <v>]                Add to existing values (repeatable)
    [--remove-value <v>]             Remove from existing values (repeatable)
    [--pass-percent N --set-percent] Update pass percentage

  gate rule remove <gate> --rule <rule-id>

EXAMPLES
  # Target an email domain in staging
  gate rule add my_gate \
    --name "Internal team" \
    --criteria email \
    --operator str_contains_any \
    --value "@company.com" \
    --env staging

  # Add a user to an existing rule
  gate rule update my_gate --rule <rule-id> --add-value "new@company.com"

  # Roll out to 10% of production
  gate rollout my_gate --percent 10 --env production

CONDITION TYPES (most common)
  email           any, none, str_contains_any, str_contains_none
  user_id         any, none, str_contains_any, regex
  country         any, none
  custom_field    any, none, gt, gte, lt, lte (use --field for the attribute)
  public          "Everyone" — no operator needed
  passes_gate     Target value is another gate's ID
  passes_segment  Target value is a segment's ID
  Run 'gate criteria' for the full list of 25 types.
`
