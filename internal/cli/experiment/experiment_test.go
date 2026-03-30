package experiment

import (
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
)

func TestFilterExperiments(t *testing.T) {
	experiments := []api.Experiment{
		{Name: "checkout_redesign", Description: "Testing new checkout flow"},
		{Name: "pricing_test", Description: "A/B test pricing tiers"},
		{Name: "onboarding_v2", Description: "New user onboarding"},
	}

	result := filterExperiments(experiments, "checkout")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}

	result = filterExperiments(experiments, "test")
	if len(result) != 2 {
		t.Errorf("expected 2 (name and description match), got %d", len(result))
	}

	result = filterExperiments(experiments, "PRICING")
	if len(result) != 1 {
		t.Errorf("case-insensitive: expected 1, got %d", len(result))
	}
}
