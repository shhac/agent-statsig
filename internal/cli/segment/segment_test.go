package segment

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

func TestFilterSegments(t *testing.T) {
	segments := []api.Segment{
		{Name: "internal_team", Description: "Internal employees"},
		{Name: "beta_users", Description: "Beta program participants"},
		{Name: "enterprise", Description: "Enterprise tier customers"},
	}

	result := filterSegments(segments, "internal")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}

	result = filterSegments(segments, "BETA")
	if len(result) != 1 {
		t.Errorf("case-insensitive: expected 1, got %d", len(result))
	}

	result = filterSegments(segments, "enterprise")
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}
