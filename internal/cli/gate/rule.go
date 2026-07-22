package gate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
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
	registerRuleMove(rule, globals)

	parent.AddCommand(rule)
}

func registerRuleList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "list <gate>",
		Short: "List rules for a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				rules, err := client.GetGateRules(ctx, args[0])
				if err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"rules": rules}, g.Format)
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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
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
				shared.WriteResource(created, g.Format)
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
		ruleRef      string
		byID         bool
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
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if ruleRef == "" {
					return agenterrors.New("--rule is required", agenterrors.FixableByAgent).
						WithHint("Use 'gate rule list <gate>' to find rule IDs")
				}

				gate, err := client.GetGate(ctx, args[0])
				if err != nil {
					return err
				}

				idx, err := shared.FindRule(gate.Rules, ruleRef, byID)
				if err != nil {
					return err
				}
				targetRule := &gate.Rules[idx]

				update := BuildRuleUpdate(targetRule, addValues, removeValues, passPercent, setPercent)
				if len(update) == 0 {
					return agenterrors.New("no updates specified", agenterrors.FixableByAgent).
						WithHint("Use --add-value, --remove-value, or --pass-percent")
				}

				if err := client.UpdateGateRule(ctx, args[0], targetRule.ID, update); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "rule": targetRule.ID}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleRef, "rule", "", "Rule ID or unique rule name")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().BoolVar(&byID, "by-id", false, "Treat --rule strictly as a rule ID (skip name resolution)")
	cmd.Flags().StringArrayVar(&addValues, "add-value", nil, "Value to add (repeatable)")
	cmd.Flags().StringArrayVar(&removeValues, "remove-value", nil, "Value to remove (repeatable)")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 0, "Pass percentage (0-100)")
	cmd.Flags().BoolVar(&setPercent, "set-percent", false, "Apply --pass-percent value")
	parent.AddCommand(cmd)
}

func registerRuleRemove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ruleRef string
	var byID bool

	cmd := &cobra.Command{
		Use:   "remove <gate>",
		Short: "Remove a rule from a gate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				resolvedID := ruleRef
				if !byID {
					gateEntity, err := client.GetGate(ctx, args[0])
					if err != nil {
						return err
					}
					idx, err := shared.FindRule(gateEntity.Rules, ruleRef, false)
					if err != nil {
						return err
					}
					resolvedID = gateEntity.Rules[idx].ID
				}
				if err := client.DeleteGateRule(ctx, args[0], resolvedID); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "deleted": resolvedID}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleRef, "rule", "", "Rule ID or unique rule name")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().BoolVar(&byID, "by-id", false, "Treat --rule strictly as a rule ID (skip name resolution and the lookup fetch)")
	parent.AddCommand(cmd)
}

func registerRuleMove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		ruleRef  string
		byID     bool
		position string
	)

	cmd := &cobra.Command{
		Use:   "move <gate>",
		Short: "Move a rule to a new position (rules evaluate top-to-bottom, first match wins)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				gateEntity, err := client.GetGate(ctx, args[0])
				if err != nil {
					return err
				}
				rules, err := shared.MoveRule(gateEntity.Rules, ruleRef, position, byID)
				if err != nil {
					return err
				}
				updated, err := client.UpdateGate(ctx, args[0], map[string]any{"rules": rules})
				if err != nil {
					return err
				}
				shared.WriteResource(updated, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleRef, "rule", "", "Rule ID or unique rule name")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().BoolVar(&byID, "by-id", false, "Treat --rule strictly as a rule ID (skip name resolution)")
	cmd.Flags().StringVar(&position, "position", "", "Target position: 1-based number, 'top', or 'bottom'")
	cmd.MarkFlagRequired("position")
	parent.AddCommand(cmd)
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
