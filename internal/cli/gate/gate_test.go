package gate

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func TestFindPublicRule(t *testing.T) {
	rules := []api.Rule{
		{ID: "r1", Name: "Email rule", Conditions: []api.Condition{{Type: "email"}}},
		{ID: "r2", Name: "Everyone", Conditions: []api.Condition{{Type: "public"}}},
	}

	id, found := FindPublicRule(rules)
	if !found {
		t.Fatal("should find public rule")
	}
	if id != "r2" {
		t.Errorf("id = %q, want r2", id)
	}

	_, found = FindPublicRule([]api.Rule{{ID: "r1", Conditions: []api.Condition{{Type: "email"}}}})
	if found {
		t.Error("should not find public rule when none exists")
	}
}

func TestFindRuleByID(t *testing.T) {
	rules := []api.Rule{
		{ID: "r1", Name: "first"},
		{ID: "r2", Name: "second"},
	}

	r := FindRuleByID(rules, "r2")
	if r == nil {
		t.Fatal("should find rule r2")
	}
	if r.Name != "second" {
		t.Errorf("name = %q", r.Name)
	}

	if FindRuleByID(rules, "r99") != nil {
		t.Error("should return nil for missing rule")
	}
}

func TestMergeConditionValues(t *testing.T) {
	existing := []string{"a@test.com", "b@test.com"}

	result := MergeConditionValues(existing, "c@test.com", "")
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}

	result = MergeConditionValues([]string{"a@test.com", "b@test.com"}, "a@test.com", "")
	if len(result) != 2 {
		t.Errorf("duplicate should not be added, got %d", len(result))
	}

	result = MergeConditionValues([]string{"a", "b", "c"}, "", "b")
	if len(result) != 2 || shared.SliceContains(result, "b") {
		t.Errorf("b should be removed: %v", result)
	}

	result = MergeConditionValues([]string{"a"}, "b,c", "a")
	if len(result) != 2 || shared.SliceContains(result, "a") {
		t.Errorf("expected [b,c], got %v", result)
	}
}

func TestBuildRuleUpdate(t *testing.T) {
	rule := &api.Rule{
		ID:         "r1",
		Conditions: []api.Condition{{Type: "email", TargetValue: []any{"a@test.com"}}},
	}

	update := BuildRuleUpdate(rule, "b@test.com", "", 0, false)
	if _, ok := update["conditions"]; !ok {
		t.Error("should have conditions in update")
	}

	update = BuildRuleUpdate(rule, "", "", 50, true)
	if update["passPercentage"] != float64(50) {
		t.Errorf("passPercentage = %v", update["passPercentage"])
	}

	update = BuildRuleUpdate(rule, "", "", 0, false)
	if len(update) != 0 {
		t.Errorf("empty update should have no keys, got %d", len(update))
	}
}
