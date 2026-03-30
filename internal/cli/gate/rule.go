package gate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/agent-statsig/internal/output"
)

func registerRule(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	rule := &cobra.Command{
		Use:   "rule",
		Short: "Manage gate rules",
	}

	registerRuleList(rule, globals)
	registerRuleAdd(rule, globals)
	registerRuleUpdate(rule, globals)
	registerRuleRemove(rule, globals)

	parent.AddCommand(rule)
}

func registerRuleList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "list <gate>",
		Short: "List rules for a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				rules, err := client.GetGateRules(ctx, args[0])
				if err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"rules": rules}, true)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
}

func registerRuleAdd(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		name         string
		criteria     string
		operator     string
		values       []string
		passPercent  float64
		environments []string
		field        string
	)

	cmd := &cobra.Command{
		Use:   "add <gate>",
		Short: "Add a rule to a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := shared.ValidateCriteria(criteria, operator); err != nil {
					return err
				}

				condition := api.Condition{
					Type:     criteria,
					Operator: operator,
				}

				if field != "" {
					condition.Field = field
				}

				if len(values) > 0 {
					condition.TargetValue = values
				}

				rule := api.Rule{
					Name:           name,
					PassPercentage: passPercent,
					Conditions:     []api.Condition{condition},
					Environments:   environments,
				}

				created, err := client.AddGateRule(ctx, args[0], rule)
				if err != nil {
					return err
				}
				output.PrintJSON(created, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&criteria, "criteria", "", "Condition type (e.g. email, user_id, country)")
	cmd.MarkFlagRequired("criteria")
	cmd.Flags().StringVar(&operator, "operator", "any", "Condition operator (default: any = case-insensitive match)")
	cmd.Flags().StringArrayVar(&values, "value", nil, "Target value (repeatable: --value a --value b)")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 100, "Pass percentage (0-100)")
	cmd.Flags().StringArrayVar(&environments, "env", nil, "Environment (repeatable: --env staging --env production)")
	cmd.Flags().StringVar(&field, "field", "", "Custom field name (for custom_field criteria)")
	parent.AddCommand(cmd)
}

func registerRuleUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		ruleID       string
		addValues    []string
		removeValues []string
		passPercent  float64
		setPercent   bool
	)

	cmd := &cobra.Command{
		Use:   "update <gate>",
		Short: "Update a rule on a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if ruleID == "" {
					return agenterrors.New("--rule is required", agenterrors.FixableByAgent).
						WithHint("Use 'gate rule list <gate>' to find rule IDs")
				}

				gate, err := client.GetGate(ctx, args[0])
				if err != nil {
					return err
				}

				targetRule := FindRuleByID(gate.Rules, ruleID)
				if targetRule == nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "rule %q not found", ruleID).
						WithHint("Use 'gate rule list " + args[0] + "' to see rule IDs")
				}

				update := BuildRuleUpdate(targetRule, addValues, removeValues, passPercent, setPercent)
				if len(update) == 0 {
					return agenterrors.New("no updates specified", agenterrors.FixableByAgent).
						WithHint("Use --add-value, --remove-value, or --pass-percent")
				}

				if err := client.UpdateGateRule(ctx, args[0], ruleID, update); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "rule": ruleID}, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleID, "rule", "", "Rule ID to update")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().StringArrayVar(&addValues, "add-value", nil, "Value to add (repeatable)")
	cmd.Flags().StringArrayVar(&removeValues, "remove-value", nil, "Value to remove (repeatable)")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 0, "Pass percentage (0-100)")
	cmd.Flags().BoolVar(&setPercent, "set-percent", false, "Apply --pass-percent value")
	parent.AddCommand(cmd)
}

func registerRuleRemove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ruleID string

	cmd := &cobra.Command{
		Use:   "remove <gate>",
		Short: "Remove a rule from a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteGateRule(ctx, args[0], ruleID); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "deleted": ruleID}, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleID, "rule", "", "Rule ID to remove")
	cmd.MarkFlagRequired("rule")
	parent.AddCommand(cmd)
}

// FindRuleByID returns the rule with the given ID, or nil.
func FindRuleByID(rules []api.Rule, id string) *api.Rule {
	for i := range rules {
		if rules[i].ID == id {
			return &rules[i]
		}
	}
	return nil
}

// BuildRuleUpdate constructs an update map for a rule, merging value changes.
func BuildRuleUpdate(rule *api.Rule, addValues, removeValues []string, passPercent float64, setPercent bool) map[string]any {
	update := make(map[string]any)
	if setPercent {
		update["passPercentage"] = passPercent
	}

	if len(addValues) == 0 && len(removeValues) == 0 {
		return update
	}

	if len(rule.Conditions) == 0 {
		return update
	}

	existing := shared.ToStringSlice(rule.Conditions[0].TargetValue)
	existing = MergeConditionValues(existing, addValues, removeValues)

	conditions := make([]api.Condition, len(rule.Conditions))
	copy(conditions, rule.Conditions)
	conditions[0].TargetValue = existing
	update["conditions"] = conditions
	return update
}

// MergeConditionValues adds and removes values from an existing slice.
func MergeConditionValues(existing, add, remove []string) []string {
	for _, v := range add {
		if !shared.SliceContains(existing, v) {
			existing = append(existing, v)
		}
	}
	for _, v := range remove {
		existing = shared.SliceRemove(existing, v)
	}
	return existing
}
