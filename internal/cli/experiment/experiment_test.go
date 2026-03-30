package experiment

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func TestFilterExperiments(t *testing.T) {
	experiments := []api.Experiment{
		{Name: "checkout_redesign", Description: "Testing new checkout flow"},
		{Name: "pricing_test", Description: "A/B test pricing tiers"},
		{Name: "onboarding_v2", Description: "New user onboarding"},
	}

	result := shared.FilterBySearch(experiments, "checkout",
		func(e api.Experiment) string { return e.Name },
		func(e api.Experiment) string { return e.Description })
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}

	result = shared.FilterBySearch(experiments, "test",
		func(e api.Experiment) string { return e.Name },
		func(e api.Experiment) string { return e.Description })
	if len(result) != 2 {
		t.Errorf("expected 2 (name and description match), got %d", len(result))
	}
}
