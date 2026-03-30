package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/config"
	"github.com/shhac/agent-statsig/internal/credential"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	"github.com/shhac/agent-statsig/internal/output"
)

type GlobalFlags struct {
	Project string
	Format  string
	Timeout int
}

func MakeContext(timeoutMs int) (context.Context, context.CancelFunc) {
	if timeoutMs > 0 {
		return context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	}
	return context.WithCancel(context.Background())
}

func ResolveProject(projectAlias string) (string, error) {
	if projectAlias != "" {
		return projectAlias, nil
	}
	if env := os.Getenv("AGENT_STATSIG_PROJECT"); env != "" {
		return env, nil
	}
	cfg := config.Read()
	if cfg.DefaultProject != "" {
		return cfg.DefaultProject, nil
	}
	available := make([]string, 0)
	for name := range cfg.Projects {
		available = append(available, name)
	}
	hint := "No projects configured. Add one with 'agent-statsig project add <alias>'"
	if len(available) > 0 {
		hint = fmt.Sprintf("Available projects: %s. Set a default with 'agent-statsig project set-default <alias>'", strings.Join(available, ", "))
	}
	return "", agenterrors.New("no project specified", agenterrors.FixableByAgent).WithHint(hint)
}

func NewClientFromProject(projectAlias string) (*api.Client, error) {
	alias, err := ResolveProject(projectAlias)
	if err != nil {
		return nil, err
	}

	cred, err := credential.Get(alias)
	if err != nil {
		var nf *credential.NotFoundError
		if errors.As(err, &nf) {
			return nil, agenterrors.Newf(agenterrors.FixableByHuman, "credentials for project %q not found", alias).
				WithHint("Add credentials with 'agent-statsig project add " + alias + " --console-key <key>'")
		}
		return nil, agenterrors.Wrap(err, agenterrors.FixableByHuman)
	}

	if cred.ConsoleKey == "" {
		return nil, agenterrors.Newf(agenterrors.FixableByHuman, "project %q has no console key", alias).
			WithHint("Update with 'agent-statsig project update " + alias + " --console-key <key>'")
	}

	return api.NewClient(cred.ConsoleKey, cred.ClientKey), nil
}

func WithClient(projectAlias string, timeout int, fn func(ctx context.Context, client *api.Client) error) error {
	ctx, cancel := MakeContext(timeout)
	defer cancel()

	client, err := NewClientFromProject(projectAlias)
	if err != nil {
		output.WriteError(os.Stderr, err)
		return nil
	}

	if err := fn(ctx, client); err != nil {
		output.WriteError(os.Stderr, err)
	}
	return nil
}

func WritePaginatedList(items []any, pagination *api.PaginationInfo, format string) {
	f := output.ResolveFormat(format)
	if f == output.FormatNDJSON {
		w := output.NewNDJSONWriter(os.Stdout)
		for _, item := range items {
			w.WriteItem(item)
		}
		if pagination != nil {
			w.WritePagination(&output.Pagination{
				HasMore:    pagination.HasMore(),
				TotalItems: pagination.TotalItems,
				Page:       pagination.PageNumber,
			})
		}
		return
	}
	result := map[string]any{"data": items}
	if pagination != nil {
		result["pagination"] = map[string]any{
			"hasMore":    pagination.HasMore(),
			"totalItems": pagination.TotalItems,
			"page":       pagination.PageNumber,
		}
	}
	output.PrintJSON(result, true)
}
