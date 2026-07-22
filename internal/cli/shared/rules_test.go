package shared

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

func namedRules(names ...string) []api.Rule {
	rules := make([]api.Rule, len(names))
	for i, n := range names {
		rules[i] = api.Rule{ID: "id-" + n, Name: n}
	}
	return rules
}

func order(rules []api.Rule) string {
	s := ""
	for _, r := range rules {
		s += r.Name
	}
	return s
}

func TestMoveRule(t *testing.T) {
	cases := []struct {
		name     string
		idOrName string
		position string
		want     string
	}{
		{"by id to top", "id-c", "top", "cab"},
		{"by id to bottom", "id-a", "bottom", "bca"},
		{"by name to middle", "c", "2", "acb"},
		{"to own position", "b", "2", "abc"},
		{"first to numeric end", "a", "3", "bca"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MoveRule(namedRules("a", "b", "c"), tc.idOrName, tc.position, false)
			if err != nil {
				t.Fatalf("MoveRule: %v", err)
			}
			if order(got) != tc.want {
				t.Errorf("order = %s, want %s", order(got), tc.want)
			}
		})
	}
}

func TestMoveRuleErrors(t *testing.T) {
	rules := namedRules("a", "b", "c")

	if _, err := MoveRule(rules, "missing", "top", false); err == nil {
		t.Error("unknown rule should error")
	}
	if _, err := MoveRule(rules, "a", "0", false); err == nil {
		t.Error("position 0 should error (1-based)")
	}
	if _, err := MoveRule(rules, "a", "4", false); err == nil {
		t.Error("out-of-range position should error")
	}

	dup := append(rules, api.Rule{ID: "id-x", Name: "a"})
	if _, err := MoveRule(dup, "a", "top", false); err == nil {
		t.Error("ambiguous name should error")
	}
	if got, err := MoveRule(dup, "id-x", "top", false); err != nil || got[0].ID != "id-x" {
		t.Errorf("ID match must win over ambiguous names: %v %v", got, err)
	}
}

func TestFindRuleByIDOnly(t *testing.T) {
	rules := namedRules("a", "b")

	if _, err := FindRule(rules, "a", true); err == nil {
		t.Error("byID must not fall back to name matching")
	}
	if idx, err := FindRule(rules, "id-b", true); err != nil || idx != 1 {
		t.Errorf("byID lookup = %d, %v", idx, err)
	}
	if idx, err := FindRule(rules, "b", false); err != nil || idx != 1 {
		t.Errorf("name lookup = %d, %v", idx, err)
	}
}
