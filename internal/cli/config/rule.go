package config

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func registerRule(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	rule := &cobra.Command{
		Use:   "rule",
		Short: "Manage dynamic config rules",
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
		Use:   "list <config>",
		Short: "List rules for a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				rules, err := client.GetConfigRules(ctx, args[0])
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
		returnValue  string
		force        bool
		dryRun       bool
	)

	cmd := &cobra.Command{
		Use:   "add <config>",
		Short: "Add a rule to a dynamic config",
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

				var rv any
				if returnValue != "" {
					if err := json.Unmarshal([]byte(returnValue), &rv); err != nil {
						return agenterrors.Newf(agenterrors.FixableByAgent, "invalid return-value JSON: %s", err)
					}
					rule.ReturnValue = rv
				}

				configEntity, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}

				if rv != nil && !force {
					if err := ValidateAgainstSchema(configEntity.Schema, rv); err != nil {
						return err
					}
				}

				rules := append(configEntity.Rules, rule)
				return applyConfigUpdate(ctx, client, g, args[0], map[string]any{"rules": rules}, dryRun)
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&criteria, "criteria", "", "Condition type")
	cmd.MarkFlagRequired("criteria")
	cmd.Flags().StringVar(&operator, "operator", "any", "Condition operator (default: any = case-insensitive match)")
	cmd.Flags().StringArrayVar(&values, "value", nil, "Target value (repeatable: --value a --value b)")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 100, "Pass percentage")
	cmd.Flags().StringArrayVar(&environments, "env", nil, "Environment (repeatable: --env staging --env production)")
	cmd.Flags().StringVar(&field, "field", "", "Custom field name")
	cmd.Flags().StringVar(&returnValue, "return-value", "", "JSON return value for this rule")
	cmd.Flags().BoolVar(&force, "force", false, "Skip client-side schema validation of --return-value")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate server-side without persisting (API dryRun)")
	parent.AddCommand(cmd)
}

func registerRuleUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		ruleRef     string
		byID        bool
		passPercent float64
		setPercent  bool
		returnValue string
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "update <config>",
		Short: "Update a rule on a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				update := make(map[string]any)
				if setPercent {
					update["passPercentage"] = passPercent
				}
				if returnValue != "" {
					var rv any
					if err := json.Unmarshal([]byte(returnValue), &rv); err != nil {
						return agenterrors.Newf(agenterrors.FixableByAgent, "invalid return-value JSON: %s", err)
					}
					update["returnValue"] = rv
				}

				if len(update) == 0 {
					return agenterrors.New("no updates specified", agenterrors.FixableByAgent).
						WithHint("Use --pass-percent or --return-value")
				}

				rv, hasRV := update["returnValue"]
				resolvedID := ruleRef
				if !byID || (hasRV && !force) {
					configEntity, err := client.GetConfig(ctx, args[0])
					if err != nil {
						return err
					}
					if !byID {
						idx, err := shared.FindRule(configEntity.Rules, ruleRef, false)
						if err != nil {
							return err
						}
						resolvedID = configEntity.Rules[idx].ID
					}
					if hasRV && !force {
						if err := ValidateAgainstSchema(configEntity.Schema, rv); err != nil {
							return err
						}
					}
				}

				if err := client.UpdateConfigRule(ctx, args[0], resolvedID, update); err != nil {
					return err
				}
				shared.WriteResource(map[string]any{"status": "ok", "rule": resolvedID}, g.Format)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleRef, "rule", "", "Rule ID or unique rule name")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().BoolVar(&byID, "by-id", false, "Treat --rule strictly as a rule ID (skip name resolution)")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 0, "Pass percentage")
	cmd.Flags().BoolVar(&setPercent, "set-percent", false, "Apply --pass-percent value")
	cmd.Flags().StringVar(&returnValue, "return-value", "", "JSON return value")
	cmd.Flags().BoolVar(&force, "force", false, "Skip client-side schema validation of --return-value")
	parent.AddCommand(cmd)
}

func registerRuleRemove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ruleRef string
	var byID bool

	cmd := &cobra.Command{
		Use:   "remove <config>",
		Short: "Remove a rule from a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				resolvedID := ruleRef
				if !byID {
					configEntity, err := client.GetConfig(ctx, args[0])
					if err != nil {
						return err
					}
					idx, err := shared.FindRule(configEntity.Rules, ruleRef, false)
					if err != nil {
						return err
					}
					resolvedID = configEntity.Rules[idx].ID
				}
				if err := client.DeleteConfigRule(ctx, args[0], resolvedID); err != nil {
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
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "move <config>",
		Short: "Move a rule to a new position (rules evaluate top-to-bottom, first match wins)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				configEntity, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}
				rules, err := shared.MoveRule(configEntity.Rules, ruleRef, position, byID)
				if err != nil {
					return err
				}
				return applyConfigUpdate(ctx, client, g, args[0], map[string]any{"rules": rules}, dryRun)
			})
		},
	}
	cmd.Flags().StringVar(&ruleRef, "rule", "", "Rule ID or unique rule name")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().BoolVar(&byID, "by-id", false, "Treat --rule strictly as a rule ID (skip name resolution)")
	cmd.Flags().StringVar(&position, "position", "", "Target position: 1-based number, 'top', or 'bottom'")
	cmd.MarkFlagRequired("position")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate server-side without persisting (API dryRun)")
	parent.AddCommand(cmd)
}
