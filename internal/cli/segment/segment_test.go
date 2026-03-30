package segment

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func TestFilterSegments(t *testing.T) {
	segments := []api.Segment{
		{Name: "internal_team", Description: "Internal employees"},
		{Name: "beta_users", Description: "Beta program participants"},
		{Name: "enterprise", Description: "Enterprise tier customers"},
	}

	result := shared.FilterBySearch(segments, "internal",
		func(s api.Segment) string { return s.Name },
		func(s api.Segment) string { return s.Description })
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}

	result = shared.FilterBySearch(segments, "BETA",
		func(s api.Segment) string { return s.Name },
		func(s api.Segment) string { return s.Description })
	if len(result) != 1 {
		t.Errorf("case-insensitive: expected 1, got %d", len(result))
	}
}
