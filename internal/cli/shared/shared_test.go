package shared

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/config"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })
}

func TestResolveProjectExplicit(t *testing.T) {
	result, err := ResolveProject("myproject")
	if err != nil {
		t.Fatal(err)
	}
	if result != "myproject" {
		t.Errorf("got %q", result)
	}
}

func TestResolveProjectEnvVar(t *testing.T) {
	t.Setenv("AGENT_STATSIG_PROJECT", "envproject")
	result, err := ResolveProject("")
	if err != nil {
		t.Fatal(err)
	}
	if result != "envproject" {
		t.Errorf("got %q", result)
	}
}

func TestResolveProjectDefault(t *testing.T) {
	setupTestDir(t)
	t.Setenv("AGENT_STATSIG_PROJECT", "")
	config.StoreProject("defaultproj", config.Project{})

	result, err := ResolveProject("")
	if err != nil {
		t.Fatal(err)
	}
	if result != "defaultproj" {
		t.Errorf("got %q", result)
	}
}

func TestResolveProjectNone(t *testing.T) {
	setupTestDir(t)
	t.Setenv("AGENT_STATSIG_PROJECT", "")

	_, err := ResolveProject("")
	if err == nil {
		t.Fatal("expected error when no project configured")
	}
}

func TestToAnySlice(t *testing.T) {
	input := []string{"a", "b", "c"}
	result := ToAnySlice(input)
	if len(result) != 3 {
		t.Fatalf("got %d", len(result))
	}
	if result[0] != "a" {
		t.Errorf("result[0] = %v", result[0])
	}
}

func TestToAnySliceEmpty(t *testing.T) {
	result := ToAnySlice([]int{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestFilterBySearch(t *testing.T) {
	type item struct{ name, desc string }
	items := []item{
		{"alpha", "First item"},
		{"beta", "Second item"},
		{"gamma", "Third alpha item"},
	}

	result := FilterBySearch(items, "alpha",
		func(i item) string { return i.name },
		func(i item) string { return i.desc })
	if len(result) != 2 {
		t.Errorf("expected 2 (name and desc match), got %d", len(result))
	}

	result = FilterBySearch(items, "BETA",
		func(i item) string { return i.name },
		func(i item) string { return i.desc })
	if len(result) != 1 {
		t.Errorf("case-insensitive: expected 1, got %d", len(result))
	}

	result = FilterBySearch(items, "nonexistent",
		func(i item) string { return i.name },
		func(i item) string { return i.desc })
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestParseJSONArg(t *testing.T) {
	result, err := ParseJSONArg(`{"key": "value"}`)
	if err != nil {
		t.Fatal(err)
	}
	if result["key"] != "value" {
		t.Errorf("key = %v", result["key"])
	}
}

func TestParseJSONArgInvalid(t *testing.T) {
	_, err := ParseJSONArg("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
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
	}

	for _, tt := range tests {
		err := ValidateCriteria(tt.criteria, tt.operator)
		if tt.wantErr && err == nil {
			t.Errorf("ValidateCriteria(%q, %q) should error", tt.criteria, tt.operator)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ValidateCriteria(%q, %q) unexpected error: %v", tt.criteria, tt.operator, err)
		}
	}
}

func TestSliceContains(t *testing.T) {
	if !SliceContains([]string{"a", "b"}, "a") {
		t.Error("should contain 'a'")
	}
	if SliceContains([]string{"a", "b"}, "c") {
		t.Error("should not contain 'c'")
	}
}

func TestSliceRemove(t *testing.T) {
	result := SliceRemove([]string{"a", "b", "c"}, "b")
	if len(result) != 2 || SliceContains(result, "b") {
		t.Errorf("'b' should be removed: %v", result)
	}
}

func TestToStringSlice(t *testing.T) {
	result := ToStringSlice([]any{"a", "b"})
	if len(result) != 2 || result[0] != "a" {
		t.Errorf("got %v", result)
	}

	result = ToStringSlice([]string{"x", "y"})
	if len(result) != 2 {
		t.Errorf("string slice: got %d", len(result))
	}

	if ToStringSlice(42) != nil {
		t.Error("non-slice should return nil")
	}
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	keys := MapKeys(m)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
