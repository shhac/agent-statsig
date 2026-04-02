package experiment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func registerUsage(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "llm-help",
		Short: "Show experiments detailed reference",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(experimentUsage)
		},
	}
	parent.AddCommand(cmd)
}

const experimentUsage = `EXPERIMENTS — Reference

An experiment is an A/B test with groups (variants). Each group has a name,
size (percentage), and parameter values. Experiments have a lifecycle:
  setup → active → decision_made (shipped) or abandoned

READ
  experiment list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  experiment get <name>              Full experiment with groups, status, hypothesis

MODIFY
  experiment create <name> [--description <text>] [--groups <json>] [--tag <tag>...]
  experiment delete <name>
  experiment archive <name>
  experiment update <name> <json>    Raw JSON partial update

LIFECYCLE
  experiment start <name>            Move from setup to active
  experiment reset <name>            Reset an active experiment
  experiment abandon <name> --reason <text>
  experiment ship <name> --group <group-id> --reason <text> [--remove-targeting]

EXAMPLES
  # Create with groups
  experiment create checkout_test \
    --description "Test new checkout flow" \
    --groups '[{"name":"control","size":50},{"name":"test","size":50,"parameterValues":{"flow":"new"}}]'

  # Ship the winning variant
  experiment ship checkout_test --group test --reason "15% CVR improvement"

GROUP FORMAT
  Groups JSON: [{"name": "control", "size": 50, "parameterValues": {...}}, ...]
  Sizes must sum to 100. parameterValues contains the config each group receives.

STATUS VALUES
  setup, active, decision_made, abandoned, archived
`
