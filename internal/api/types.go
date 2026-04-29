package api

import "encoding/json"

// Gate represents a Statsig feature gate.
type Gate struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	IsEnabled   bool            `json:"isEnabled"`
	Tags        []string        `json:"tags,omitempty"`
	Rules       []Rule          `json:"rules,omitempty"`
	Salt        string          `json:"salt,omitempty"`
	Status      string          `json:"status,omitempty"`
	Type        string          `json:"type,omitempty"`
	TargetApps  json.RawMessage `json:"targetApps,omitempty"`
	CreatorID   string          `json:"creatorID,omitempty"`
}

// DynamicConfig represents a Statsig dynamic config.
type DynamicConfig struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	IsEnabled    bool            `json:"isEnabled"`
	Tags         []string        `json:"tags,omitempty"`
	Rules        []Rule          `json:"rules,omitempty"`
	DefaultValue json.RawMessage `json:"defaultValue,omitempty"`
	Schema       json.RawMessage `json:"schema,omitempty"`
	Salt         string          `json:"salt,omitempty"`
	Status       string          `json:"status,omitempty"`
	Type         string          `json:"type,omitempty"`
	CreatorID    string          `json:"creatorID,omitempty"`
}

// Experiment represents a Statsig experiment.
type Experiment struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	IsEnabled   bool            `json:"isEnabled"`
	Tags        []string        `json:"tags,omitempty"`
	Status      string          `json:"status,omitempty"`
	Hypothesis  string          `json:"hypothesis,omitempty"`
	Groups      []Group         `json:"groups,omitempty"`
	Rules       []Rule          `json:"rules,omitempty"`
	Salt        string          `json:"salt,omitempty"`
	Type        string          `json:"type,omitempty"`
	LayerID     string          `json:"layerID,omitempty"`
	TargetApps  json.RawMessage `json:"targetApps,omitempty"`
	CreatorID   string          `json:"creatorID,omitempty"`
}

// Segment represents a Statsig segment.
type Segment struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Rules       []Rule   `json:"rules,omitempty"`
	CreatorID   string   `json:"creatorID,omitempty"`
}

// Rule represents a targeting rule shared by gates, configs, and experiments.
type Rule struct {
	ID             string      `json:"id,omitempty"`
	Name           string      `json:"name"`
	PassPercentage float64     `json:"passPercentage"`
	Conditions     []Condition `json:"conditions"`
	ReturnValue    any         `json:"returnValue,omitempty"`
	Environments   []string    `json:"environments,omitempty"`
}

// Condition represents a single condition within a rule.
type Condition struct {
	Type        string `json:"type"`
	Operator    string `json:"operator,omitempty"`
	TargetValue any    `json:"targetValue,omitempty"`
	Field       string `json:"field,omitempty"`
	CustomID    string `json:"customID,omitempty"`
}

// Group represents an experiment group (variant).
type Group struct {
	Name            string         `json:"name"`
	Size            float64        `json:"size"`
	ParameterValues map[string]any `json:"parameterValues,omitempty"`
}

// Tag represents a Statsig tag.
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsCore      bool   `json:"isCore"`
}

// ConditionTypes lists all known Statsig condition types.
var ConditionTypes = []string{
	"app_version", "browser_name", "browser_version", "country",
	"custom_field", "device_model", "email", "environment_tier",
	"experiment_group", "fails_gate", "fails_segment", "ip_address",
	"javascript", "locale", "os_name", "os_version",
	"passes_gate", "passes_segment", "public", "target_app",
	"time", "unit_id", "url", "user_agent", "user_id",
}

// OperatorsByType maps condition types to their valid operators.
var OperatorsByType = map[string][]string{
	"user_id":          {"any", "none", "is_null", "is_not_null", "str_contains_any", "str_contains_none", "regex"},
	"email":            {"any", "none", "str_contains_any", "str_contains_none"},
	"country":          {"any", "none"},
	"ip_address":       {"any", "none"},
	"locale":           {"any", "none"},
	"app_version":      {"version_gt", "version_gte", "version_lt", "version_lte", "any", "none"},
	"browser_version":  {"version_gt", "version_gte", "version_lt", "version_lte", "any", "none"},
	"os_version":       {"version_gt", "version_gte", "version_lt", "version_lte", "any", "none"},
	"browser_name":     {"any", "none"},
	"os_name":          {"any", "none"},
	"device_model":     {"any", "none", "is_null", "is_not_null", "str_contains_any", "str_contains_none", "regex"},
	"unit_id":          {"any", "none", "is_null", "is_not_null", "str_contains_any", "str_contains_none", "regex"},
	"custom_field":     {"any", "none", "str_contains_any", "str_contains_none", "gt", "gte", "lt", "lte", "version_gt", "version_gte", "version_lt", "version_lte", "before", "after"},
	"time":             {"after", "before"},
	"environment_tier": {},
	"passes_gate":      {},
	"fails_gate":       {},
	"passes_segment":   {},
	"fails_segment":    {},
	"public":           {},
}
