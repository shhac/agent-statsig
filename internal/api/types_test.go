package api

import (
	"encoding/json"
	"testing"
)

func TestGateMarshal(t *testing.T) {
	gate := Gate{
		ID:        "gate-123",
		Name:      "test_gate",
		IsEnabled: true,
		Rules: []Rule{
			{
				Name:           "Email rule",
				PassPercentage: 100,
				Conditions: []Condition{
					{Type: "email", Operator: "any", TargetValue: []string{"user@test.com"}},
				},
			},
		},
	}

	b, err := json.Marshal(gate)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Gate
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Name != "test_gate" {
		t.Errorf("Name = %q", decoded.Name)
	}
	if !decoded.IsEnabled {
		t.Error("IsEnabled should be true")
	}
	if len(decoded.Rules) != 1 {
		t.Fatalf("got %d rules", len(decoded.Rules))
	}
	if decoded.Rules[0].Conditions[0].Type != "email" {
		t.Errorf("condition type = %q", decoded.Rules[0].Conditions[0].Type)
	}
}

func TestExperimentGroupMarshal(t *testing.T) {
	exp := Experiment{
		Name: "test_exp",
		Groups: []Group{
			{Name: "control", Size: 50},
			{Name: "test", Size: 50, ParameterValues: map[string]any{"color": "blue"}},
		},
	}

	b, err := json.Marshal(exp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Experiment
	json.Unmarshal(b, &decoded)

	if len(decoded.Groups) != 2 {
		t.Fatalf("got %d groups", len(decoded.Groups))
	}
	if decoded.Groups[1].ParameterValues["color"] != "blue" {
		t.Errorf("color = %v", decoded.Groups[1].ParameterValues["color"])
	}
}

func TestRuleWithReturnValue(t *testing.T) {
	rule := Rule{
		Name:           "Config rule",
		PassPercentage: 100,
		Conditions:     []Condition{{Type: "public"}},
		ReturnValue:    map[string]any{"theme": "dark", "limit": 50},
	}

	b, _ := json.Marshal(rule)
	var decoded Rule
	json.Unmarshal(b, &decoded)

	rv, ok := decoded.ReturnValue.(map[string]any)
	if !ok {
		t.Fatal("ReturnValue should be a map")
	}
	if rv["theme"] != "dark" {
		t.Errorf("theme = %v", rv["theme"])
	}
}

func TestConditionWithCustomField(t *testing.T) {
	cond := Condition{
		Type:        "custom_field",
		Operator:    "gt",
		TargetValue: 31,
		Field:       "age",
	}

	b, _ := json.Marshal(cond)
	var decoded Condition
	json.Unmarshal(b, &decoded)

	if decoded.Field != "age" {
		t.Errorf("Field = %q", decoded.Field)
	}
	if decoded.Operator != "gt" {
		t.Errorf("Operator = %q", decoded.Operator)
	}
}

func TestSegmentTypes(t *testing.T) {
	seg := Segment{
		Name: "internal_team",
		Type: "id_list",
	}
	b, _ := json.Marshal(seg)
	var decoded Segment
	json.Unmarshal(b, &decoded)
	if decoded.Type != "id_list" {
		t.Errorf("Type = %q", decoded.Type)
	}
}
