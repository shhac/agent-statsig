package config

import (
	"encoding/json"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestDryRunDoesNotPersist(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})
	out, stderr := runConfigCmd(t, srv.Handler(), "config", "value", "set", "my_config", `{"theme":"dark"}`, "--dry-run")

	if stderr != "" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["dryRun"] != true {
		t.Errorf("output should be marked dryRun: %s", out)
	}
	if srv.PatchCount() != 0 {
		t.Error("dry-run must not persist a PATCH")
	}
	if srv.DryRunPatchCount() != 1 {
		t.Errorf("DryRunPatchCount = %d", srv.DryRunPatchCount())
	}
}

func TestDryRunOnUpdateAndRuleAdd(t *testing.T) {
	srv := mockstatsig.NewConfigServer(api.DynamicConfig{Name: "my_config"})

	_, stderr := runConfigCmd(t, srv.Handler(), "config", "update", "my_config", `{"description":"x"}`, "--dry-run")
	if stderr != "" {
		t.Fatalf("update --dry-run: %s", stderr)
	}

	_, stderr = runConfigCmd(t, srv.Handler(), "config", "rule", "add", "my_config",
		"--name", "R", "--criteria", "public", "--return-value", `{"a":1}`, "--dry-run")
	if stderr != "" {
		t.Fatalf("rule add --dry-run: %s", stderr)
	}

	if srv.PatchCount() != 0 {
		t.Error("dry-run must not persist")
	}
	if srv.DryRunPatchCount() != 2 {
		t.Errorf("DryRunPatchCount = %d", srv.DryRunPatchCount())
	}
}
