package config

import (
	"context"
	"encoding/json"
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
		Short: "Manage dynamic config rules",
	}

	registerRuleList(rule, globals)
	registerRuleAdd(rule, globals)
	registerRuleUpdate(rule, globals)
	registerRuleRemove(rule, globals)

	parent.AddCommand(rule)
}

func registerRuleList(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	cmd := &cobra.Command{
		Use:   "list <config>",
		Short: "List rules for a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				rules, err := client.GetConfigRules(ctx, args[0])
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
		returnValue  string
	)

	cmd := &cobra.Command{
		Use:   "add <config>",
		Short: "Add a rule to a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
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

				if returnValue != "" {
					var rv any
					if err := json.Unmarshal([]byte(returnValue), &rv); err != nil {
						return agenterrors.Newf(agenterrors.FixableByAgent, "invalid return-value JSON: %s", err)
					}
					rule.ReturnValue = rv
				}

				configEntity, err := client.GetConfig(ctx, args[0])
				if err != nil {
					return err
				}

				if returnValue != "" && configEntity.Schema != nil {
					var rv any
					json.Unmarshal([]byte(returnValue), &rv)
					if err := validateAgainstSchema(configEntity.Schema, rv); err != nil {
						return err
					}
				}

				rules := append(configEntity.Rules, rule)
				update := map[string]any{"rules": rules}
				updated, err := client.UpdateConfig(ctx, args[0], update)
				if err != nil {
					return err
				}
				output.PrintJSON(updated, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Rule name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&criteria, "criteria", "", "Condition type")
	cmd.MarkFlagRequired("criteria")
	cmd.Flags().StringVar(&operator, "operator", "", "Condition operator")
	cmd.Flags().StringVar(&values, "values", "", "Comma-separated target values")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 100, "Pass percentage")
	cmd.Flags().StringVar(&environments, "environments", "", "Comma-separated environments")
	cmd.Flags().StringVar(&field, "field", "", "Custom field name")
	cmd.Flags().StringVar(&returnValue, "return-value", "", "JSON return value for this rule")
	parent.AddCommand(cmd)
}

func registerRuleUpdate(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var (
		ruleID      string
		passPercent float64
		setPercent  bool
		returnValue string
	)

	cmd := &cobra.Command{
		Use:   "update <config>",
		Short: "Update a rule on a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
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

				if err := client.UpdateConfigRule(ctx, args[0], ruleID, update); err != nil {
					return err
				}
				output.PrintJSON(map[string]any{"status": "ok", "rule": ruleID}, true)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&ruleID, "rule", "", "Rule ID to update")
	cmd.MarkFlagRequired("rule")
	cmd.Flags().Float64Var(&passPercent, "pass-percent", 0, "Pass percentage")
	cmd.Flags().BoolVar(&setPercent, "set-percent", false, "Apply --pass-percent value")
	cmd.Flags().StringVar(&returnValue, "return-value", "", "JSON return value")
	parent.AddCommand(cmd)
}

func registerRuleRemove(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var ruleID string

	cmd := &cobra.Command{
		Use:   "remove <config>",
		Short: "Remove a rule from a dynamic config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				if err := client.DeleteConfigRule(ctx, args[0], ruleID); err != nil {
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

func validateAgainstSchema(schema json.RawMessage, value any) error {
	var schemaObj map[string]any
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		return nil
	}

	props, ok := schemaObj["properties"].(map[string]any)
	if !ok {
		return nil
	}

	valueMap, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	required, _ := schemaObj["required"].([]any)
	requiredSet := make(map[string]bool)
	for _, r := range required {
		if s, ok := r.(string); ok {
			requiredSet[s] = true
		}
	}

	for key := range requiredSet {
		if _, ok := valueMap[key]; !ok {
			return agenterrors.Newf(agenterrors.FixableByAgent, "missing required field %q in return value", key).
				WithHint("Schema requires: " + strings.Join(mapKeys(requiredSet), ", "))
		}
	}

	for key := range valueMap {
		if _, ok := props[key]; !ok {
			knownKeys := mapKeys(props)
			return agenterrors.Newf(agenterrors.FixableByAgent, "unknown field %q in return value", key).
				WithHint("Known fields: " + strings.Join(knownKeys, ", "))
		}
	}

	return nil
}

func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
