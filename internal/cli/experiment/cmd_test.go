package experiment

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/clitest"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestExperimentList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.List([]api.Experiment{{Name: "exp1"}, {Name: "exp2"}}, 2))
	}, "experiment", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestExperimentGet(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity(api.Experiment{Name: "my_exp", Status: "active"}))
	}, "experiment", "get", "my_exp")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "active" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestExperimentCreate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "new_exp" {
			t.Errorf("name = %v", body["name"])
		}
		w.Write(mockstatsig.Entity(api.Experiment{Name: "new_exp"}))
	}, "experiment", "create", "new_exp", "--description", "Test experiment")

	if out == "" {
		t.Error("expected output")
	}
}

func TestExperimentCreateWithGroups(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		groups := body["groups"].([]any)
		if len(groups) != 2 {
			t.Errorf("expected 2 groups, got %d", len(groups))
		}
		w.Write(mockstatsig.Entity(api.Experiment{Name: "new_exp"}))
	}, "experiment", "create", "new_exp",
		"--groups", `[{"name":"control","size":50},{"name":"test","size":50}]`)

	if out == "" {
		t.Error("expected output")
	}
}

func TestExperimentStart(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/experiments/my_exp/start" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "experiment", "start", "my_exp")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["started"] != true {
		t.Errorf("started = %v", parsed["started"])
	}
}

func TestExperimentAbandon(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["decisionReason"] != "not useful" {
			t.Errorf("reason = %v", body["decisionReason"])
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "experiment", "abandon", "my_exp", "--reason", "not useful")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["abandoned"] != true {
		t.Errorf("abandoned = %v", parsed["abandoned"])
	}
}

func TestExperimentShip(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["id"] != "group-1" {
			t.Errorf("group = %v", body["id"])
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "experiment", "ship", "my_exp", "--group", "group-1", "--reason", "winner")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["shipped"] != "group-1" {
		t.Errorf("shipped = %v", parsed["shipped"])
	}
}

func TestExperimentDelete(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "experiment", "delete", "old_exp")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestExperimentUpdate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity(api.Experiment{Name: "updated"}))
	}, "experiment", "update", "my_exp", `{"description":"new"}`)

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "updated" {
		t.Errorf("name = %v", parsed["name"])
	}
}
