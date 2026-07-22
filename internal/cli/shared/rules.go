package shared

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shhac/agent-statsig/internal/api"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

// FindRule resolves a rule reference to its index. An exact ID match wins;
// otherwise a unique rule name matches, unless byID restricts resolution to
// IDs. An ambiguous name errors and lists the candidate IDs so callers can
// retry with one.
func FindRule(rules []api.Rule, ref string, byID bool) (int, error) {
	for i, r := range rules {
		if r.ID == ref {
			return i, nil
		}
	}
	if byID {
		return 0, agenterrors.Newf(agenterrors.FixableByAgent, "rule with ID %q not found", ref).
			WithHint("Use 'rule list' to see rule IDs, or drop --by-id to match by name")
	}

	idx := -1
	var ids []string
	for i, r := range rules {
		if r.Name != ref {
			continue
		}
		ids = append(ids, r.ID)
		if idx == -1 {
			idx = i
		}
	}
	if len(ids) > 1 {
		return 0, agenterrors.Newf(agenterrors.FixableByAgent, "multiple rules named %q", ref).
			WithHint(fmt.Sprintf("Matching rule IDs: %s — pass --rule <id> instead", strings.Join(ids, ", ")))
	}
	if idx == -1 {
		return 0, agenterrors.Newf(agenterrors.FixableByAgent, "rule %q not found", ref).
			WithHint("Use 'rule list' to see rule IDs and names")
	}
	return idx, nil
}

// MoveRule returns a reordered copy of rules with the rule matching ref
// (resolved via FindRule) moved to a 1-based position; "top" and "bottom" are
// also accepted. Rules evaluate top-to-bottom, so position 1 wins first.
func MoveRule(rules []api.Rule, ref, position string, byID bool) ([]api.Rule, error) {
	idx, err := FindRule(rules, ref, byID)
	if err != nil {
		return nil, err
	}

	target, err := resolvePosition(position, len(rules))
	if err != nil {
		return nil, err
	}

	moved := rules[idx]
	rest := make([]api.Rule, 0, len(rules)-1)
	rest = append(rest, rules[:idx]...)
	rest = append(rest, rules[idx+1:]...)

	out := make([]api.Rule, 0, len(rules))
	out = append(out, rest[:target]...)
	out = append(out, moved)
	out = append(out, rest[target:]...)
	return out, nil
}

func resolvePosition(position string, count int) (int, error) {
	switch position {
	case "top":
		return 0, nil
	case "bottom":
		return count - 1, nil
	}
	n, err := strconv.Atoi(position)
	if err != nil || n < 1 || n > count {
		return 0, agenterrors.Newf(agenterrors.FixableByAgent, "invalid position %q", position).
			WithHint(fmt.Sprintf("Use a number from 1 to %d, 'top', or 'bottom'", count))
	}
	return n - 1, nil
}
