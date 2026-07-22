package shared

import (
	"fmt"
	"strconv"

	"github.com/shhac/agent-statsig/internal/api"
	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

// MoveRule returns a reordered copy of rules with the rule matching idOrName
// (matched by ID first, then by unique name) moved to a 1-based position;
// "top" and "bottom" are also accepted. Rules evaluate top-to-bottom, so
// position 1 wins first.
func MoveRule(rules []api.Rule, idOrName, position string) ([]api.Rule, error) {
	idx := findRuleIndex(rules, idOrName)
	if idx == -2 {
		return nil, agenterrors.Newf(agenterrors.FixableByAgent, "multiple rules named %q", idOrName).
			WithHint("Use the rule ID from 'rule list' instead of the name")
	}
	if idx == -1 {
		return nil, agenterrors.Newf(agenterrors.FixableByAgent, "rule %q not found", idOrName).
			WithHint("Use 'rule list' to see rule IDs and names")
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

// findRuleIndex matches by ID first, then by name. Returns -1 when not found
// and -2 when the name is ambiguous.
func findRuleIndex(rules []api.Rule, idOrName string) int {
	for i, r := range rules {
		if r.ID == idOrName {
			return i
		}
	}
	idx := -1
	for i, r := range rules {
		if r.Name != idOrName {
			continue
		}
		if idx != -1 {
			return -2
		}
		idx = i
	}
	return idx
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
