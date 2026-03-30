package gate

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
	"github.com/shhac/agent-statsig/internal/output"
)

// FindPublicRule returns the ID of the first rule with a "public" condition, or empty string if none.
func FindPublicRule(rules []api.Rule) (string, bool) {
	for _, r := range rules {
		for _, c := range r.Conditions {
			if c.Type == "public" {
				return r.ID, true
			}
		}
	}
	return "", false
}

func registerRollout(parent *cobra.Command, globals func() *shared.GlobalFlags) {
	var percent float64
	var environments string

	cmd := &cobra.Command{
		Use:   "rollout <name>",
		Short: "Set rollout percentage (Everyone rule)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return shared.WithClient(g.Project, g.Timeout, func(ctx context.Context, client *api.Client) error {
				gate, err := client.GetGate(ctx, args[0])
				if err != nil {
					return err
				}

				var envs []string
				if environments != "" {
					envs = strings.Split(environments, ",")
				}

				if ruleID, found := FindPublicRule(gate.Rules); found {
					update := map[string]any{"passPercentage": percent}
					if len(envs) > 0 {
						update["environments"] = envs
					}
					if err := client.UpdateGateRule(ctx, args[0], ruleID, update); err != nil {
						return err
					}
				} else {
					rule := api.Rule{
						Name:           "Everyone",
						PassPercentage: percent,
						Conditions:     []api.Condition{{Type: "public"}},
						Environments:   envs,
					}
					if _, err := client.AddGateRule(ctx, args[0], rule); err != nil {
						return err
					}
				}

				output.PrintJSON(map[string]any{
					"status":         "ok",
					"gate":           args[0],
					"rolloutPercent": percent,
				}, true)
				return nil
			})
		},
	}
	cmd.Flags().Float64Var(&percent, "percent", 0, "Rollout percentage (0-100)")
	cmd.MarkFlagRequired("percent")
	cmd.Flags().StringVar(&environments, "environments", "", "Comma-separated environments")
	parent.AddCommand(cmd)
}
