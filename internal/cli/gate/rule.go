package gate

import (
	"context"
	"strings"

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
		values       string
		passPercent  float64
		environments string
		field        string
	)

	cmd := &cobra.Command{
		Use:   "add <gate>",
		Short: "Add a rule to a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := validateCriteria(criteria, operator); err != nil {
					return err
				}

				condition := api.Condition{
					Type:     criteria,
					Operator: operator,
				}

				if field != "" {
					condition.Field = field
				}

				if values != "" {
					condition.TargetValue = strings.Split(values, ",")
				}

				rule := api.Rule{
					Name:           name,
					PassPercentage: passPercent,
					Conditions:     []api.Condition{condition},
				}

				if environments != "" {
					rule.Environments = strings.Split(environments, ",")
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
	cmd.Flags().StringVar(&operator, "operator", "", "Condition operator (e.g. any, none, str_contains_any)")
	cmd.Flags().StringVar(&values, "values", "", "Comma-separated target values")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 100, "Pass percentage (0-100)")
	cmd.Flags().StringVar(&environments, "environments", "", "Comma-separated environments")
	cmd.Flags().StringVar(&field, "field", "", "Custom field name (for custom_field criteria)")
	parent.AddCommand(cmd)
}

func registerRuleUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		ruleID       string
		addValues    string
		removeValues string
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

				var targetRule *api.Rule
				for _, r := range gate.Rules {
					if r.ID == ruleID {
						targetRule = &r
						break
					}
				}
				if targetRule == nil {
					return agenterrors.Newf(agenterrors.FixableByAgent, "rule %q not found", ruleID).
						WithHint("Use 'gate rule list " + args[0] + "' to see rule IDs")
				}

				update := make(map[string]any)
				if setPercent {
					update["passPercentage"] = passPercent
				}

				if addValues != "" || removeValues != "" {
					if len(targetRule.Conditions) == 0 {
						return agenterrors.New("rule has no conditions to modify values on", agenterrors.FixableByAgent)
					}

					existing := targetRule.Conditions[0]
					existingVals := toStringSlice(existing.TargetValue)

					if addValues != "" {
						for _, v := range strings.Split(addValues, ",") {
							if !contains(existingVals, v) {
								existingVals = append(existingVals, v)
							}
						}
					}
					if removeValues != "" {
						for _, v := range strings.Split(removeValues, ",") {
							existingVals = removeFromSlice(existingVals, v)
						}
					}

					conditions := make([]api.Condition, len(targetRule.Conditions))
					copy(conditions, targetRule.Conditions)
					conditions[0].TargetValue = existingVals
					update["conditions"] = conditions
				}

				if len(update) == 0 {
					return agenterrors.New("no updates specified", agenterrors.FixableByAgent).
						WithHint("Use --add-values, --remove-values, or --pass-percent")
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
	cmd.Flags().StringVar(&addValues, "add-values", "", "Values to add (comma-separated)")
	cmd.Flags().StringVar(&removeValues, "remove-values", "", "Values to remove (comma-separated)")
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

func validateCriteria(criteria, operator string) error {
	found := false
	for _, ct := range api.ConditionTypes {
		if ct == criteria {
			found = true
			break
		}
	}
	if !found {
		return agenterrors.Newf(agenterrors.FixableByAgent, "unknown criteria %q", criteria).
			WithHint("Use 'gate criteria' to list available criteria types")
	}

	if operator == "" {
		return nil
	}

	ops, ok := api.OperatorsByType[criteria]
	if !ok || len(ops) == 0 {
		return nil
	}

	for _, op := range ops {
		if op == operator {
			return nil
		}
	}
	return agenterrors.Newf(agenterrors.FixableByAgent, "invalid operator %q for criteria %q", operator, criteria).
		WithHint("Valid operators: " + strings.Join(ops, ", "))
}

func toStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	default:
		return nil
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
