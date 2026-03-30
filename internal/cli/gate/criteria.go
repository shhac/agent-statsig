package gate

import (
	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/output"
)

func registerCriteria(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "criteria",
		Short: "List available condition criteria and operators",
		Run: func(cmd *cobra.Command, args []string) {
			criteria := make([]map[string]any, 0, len(api.ConditionTypes))
			for _, ct := range api.ConditionTypes {
				entry := map[string]any{"type": ct}
				if ops, ok := api.OperatorsByType[ct]; ok && len(ops) > 0 {
					entry["operators"] = ops
				}
				criteria = append(criteria, entry)
			}
			output.PrintJSON(map[string]any{"criteria": criteria}, true)
		},
	}
	parent.AddCommand(cmd)
}
