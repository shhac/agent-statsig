package segment

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/clitest"
	"github.com/shhac/agent-statsig/internal/mockstatsig"
)

func TestSegmentList(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.List([]api.Segment{{Name: "seg1"}, {Name: "seg2"}}, 2))
	}, "segment", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentGet(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity(api.Segment{Name: "internal_team", Type: "id_list"}))
	}, "segment", "get", "internal_team")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "internal_team" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestSegmentCreate(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(mockstatsig.Entity(api.Segment{Name: "new_seg"}))
	}, "segment", "create", "new_seg", "--type", "id_list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentDelete(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "delete", "old_seg")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestSegmentArchive(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "archive", "old_seg")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["archived"] != true {
		t.Errorf("archived = %v", parsed["archived"])
	}
}

func TestSegmentIDsGet(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockstatsig.Entity([]string{"user1", "user2"}))
	}, "segment", "ids", "get", "my_seg")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentIDsAdd(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		ids := body["ids"].([]any)
		if len(ids) != 2 {
			t.Errorf("expected 2 ids, got %d", len(ids))
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "ids", "add", "my_seg", "--id", "user1", "--id", "user2")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["added"] != float64(2) {
		t.Errorf("added = %v", parsed["added"])
	}
}

func TestSegmentIDsRemove(t *testing.T) {
	out, _ := clitest.Run(t, Register, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "ids", "remove", "my_seg", "--id", "user1")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["removed"] != float64(1) {
		t.Errorf("removed = %v", parsed["removed"])
	}
}
