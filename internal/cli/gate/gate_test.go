package gate

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

func TestFilterGates(t *testing.T) {
	gates := []api.Gate{
		{Name: "feature_onboarding", Description: "Onboarding flow"},
		{Name: "feature_checkout", Description: "Checkout v2"},
		{Name: "internal_debug", Description: "Debug tools"},
	}

	result := filterGates(gates, "onboarding")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Name != "feature_onboarding" {
		t.Errorf("name = %q", result[0].Name)
	}

	result = filterGates(gates, "FEATURE")
	if len(result) != 2 {
		t.Errorf("case-insensitive search should match 2, got %d", len(result))
	}

	result = filterGates(gates, "nonexistent")
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestValidateCriteria(t *testing.T) {
	tests := []struct {
		criteria string
		operator string
		wantErr  bool
	}{
		{"email", "any", false},
		{"email", "str_contains_any", false},
		{"email", "gt", true},
		{"user_id", "any", false},
		{"unknown_type", "", true},
		{"public", "", false},
		{"environment_tier", "", false},
		{"custom_field", "gt", false},
		{"custom_field", "any", false},
	}

	for _, tt := range tests {
		err := validateCriteria(tt.criteria, tt.operator)
		if tt.wantErr && err == nil {
			t.Errorf("validateCriteria(%q, %q) should error", tt.criteria, tt.operator)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("validateCriteria(%q, %q) unexpected error: %v", tt.criteria, tt.operator, err)
		}
	}
}

func TestToStringSlice(t *testing.T) {
	result := toStringSlice([]any{"a", "b", "c"})
	if len(result) != 3 {
		t.Fatalf("got %d items", len(result))
	}
	if result[0] != "a" || result[2] != "c" {
		t.Errorf("unexpected values: %v", result)
	}

	result = toStringSlice([]string{"x", "y"})
	if len(result) != 2 {
		t.Errorf("string slice: got %d", len(result))
	}

	result = toStringSlice(42)
	if result != nil {
		t.Errorf("non-slice should return nil, got %v", result)
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b"}, "a") {
		t.Error("should contain 'a'")
	}
	if contains([]string{"a", "b"}, "c") {
		t.Error("should not contain 'c'")
	}
}

func TestRemoveFromSlice(t *testing.T) {
	result := removeFromSlice([]string{"a", "b", "c"}, "b")
	if len(result) != 2 {
		t.Fatalf("got %d items", len(result))
	}
	if contains(result, "b") {
		t.Error("'b' should be removed")
	}
}
