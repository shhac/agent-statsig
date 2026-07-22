package shared

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	libcli "github.com/shhac/lib-agent-cli/cli"
)

// GetEntities runs the family's multi-capable get: sets up one client then
// resolves each id through getOne and streams per the shared get contract
// (NDJSON by default; item-level misses become @unresolved records on stdout;
// command-level failures bubble to the caller).
func GetEntities(g *GlobalFlags, args []string, getOne func(ctx context.Context, client *api.Client, id string) (any, error)) error {
	return WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
		return libcli.EntityGet(os.Stdout, g.Format, args, func(id string) (any, error) {
			return getOne(ctx, client, id)
		})
	})
}

// RegisterAction wires a simple single-id mutation command: ExactArgs(1) +
// WithClient + one client call, emitting the standard
// {"status":"ok", resultKey: id} payload. Pass extra to add fixed fields
// alongside the result key (e.g. {"isEnabled": true}); preserve each command's
// exact payload by matching its original keys.
func RegisterAction(parent *cobra.Command, globals func() *GlobalFlags, use, short, resultKey string, extra map[string]any, call func(ctx context.Context, client *api.Client, id string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := globals()
			return WithClient(g.Project, g.TimeoutMS, g.Debug, func(ctx context.Context, client *api.Client) error {
				if err := call(ctx, client, args[0]); err != nil {
					return err
				}
				payload := map[string]any{"status": "ok", resultKey: args[0]}
				for k, v := range extra {
					payload[k] = v
				}
				WriteResource(payload, g.Format)
				return nil
			})
		},
	}
	parent.AddCommand(cmd)
	return cmd
}
