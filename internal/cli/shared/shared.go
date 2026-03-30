package shared

import (
	"context"
	"encoding/json"
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

// ToAnySlice converts a typed slice to []any.
func ToAnySlice[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// FilterBySearch filters a slice by substring match on name/description fields.
func FilterBySearch[T any](items []T, search string, getName func(T) string, getDesc func(T) string) []T {
	search = strings.ToLower(search)
	var filtered []T
	for _, item := range items {
		if strings.Contains(strings.ToLower(getName(item)), search) ||
			strings.Contains(strings.ToLower(getDesc(item)), search) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// ParseJSONArg parses a JSON string argument into a map, returning a classified error on failure.
func ParseJSONArg(raw string) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, agenterrors.Newf(agenterrors.FixableByAgent, "invalid JSON: %s", err).
			WithHint("Provide a valid JSON object")
	}
	return result, nil
}

// ValidateCriteria checks that a criteria type and operator are valid Statsig condition values.
func ValidateCriteria(criteria, operator string) error {
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

// Slice helpers

func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func SliceRemove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func ToStringSlice(v any) []string {
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

func MapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
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
