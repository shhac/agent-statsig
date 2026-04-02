package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func registerUsage(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "llm-help",
		Short: "Show dynamic configs detailed reference",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(configUsage)
		},
	}
	parent.AddCommand(cmd)
}

const configUsage = `DYNAMIC CONFIGS — Reference

A dynamic config returns a JSON object whose value varies by targeting rules.
Like gates, rules are evaluated top-to-bottom and use the same condition types.
Unlike gates, each rule has a returnValue (the JSON payload for matching users).

Configs may have a JSON Schema. When present, return values are validated
client-side before the API call — type errors, missing fields, and unknown
fields are caught with helpful hints.

READ
  config list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  config get <name>                  Full config with rules, values, and schema
  config rule list <config>          List rules with their IDs

MODIFY
  config create <name> [--description <text>] [--tag <tag>...]
  config delete <name>
  config enable <name>
  config disable <name>
  config archive <name>
  config update <name> <json>        Raw JSON partial update (escape hatch)

RULE MANAGEMENT
  config rule add <config>
    --name <rule-name>
    --criteria <type>
    [--operator <op>]                Default: any
    [--value <v>]                    Repeatable
    --return-value <json>            The JSON value this rule returns
    [--pass-percent N]               Default: 100
    [--env <environment>]            Repeatable
    [--field <name>]                 For custom_field criteria

  config rule update <config>
    --rule <rule-id>
    [--pass-percent N --set-percent]
    [--return-value <json>]

  config rule remove <config> --rule <rule-id>

EXAMPLES
  # Add a rule returning custom values for internal users
  config rule add my_config \
    --name "Internal" \
    --criteria email \
    --operator str_contains_any \
    --value "@company.com" \
    --return-value '{"enabledGlobally": true, "allowOrganizers": []}' \
    --env staging

SCHEMA VALIDATION
  If the config has a JSON Schema, --return-value is validated before sending:
    ✗ Missing required field → "Schema requires: field1, field2"
    ✗ Unknown field          → "Known fields: field1, field2"
    ✗ Wrong type             → "expected boolean, got string"
  All errors are fixable_by: agent with hints.
`
