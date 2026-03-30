package segment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/cli/shared"
)

func globals() *shared.GlobalFlags {
	return &shared.GlobalFlags{}
}

func runSegmentCmd(t *testing.T, handler http.HandlerFunc, args ...string) (string, string) {
	t.Helper()
	shared.SetupMockServer(t, handler)

	root := &cobra.Command{Use: "test"}
	Register(root, func() *shared.GlobalFlags { return globals() })

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	root.SetArgs(args)
	root.Execute()

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var outBuf, errBuf bytes.Buffer
	outBuf.ReadFrom(rOut)
	errBuf.ReadFrom(rErr)

	return outBuf.String(), errBuf.String()
}

func entityJSON(data any) []byte {
	b, _ := json.Marshal(map[string]any{"data": data})
	return b
}

func listJSON(data any, total int) []byte {
	b, _ := json.Marshal(map[string]any{
		"data":       data,
		"pagination": map[string]any{"itemsPerPage": 100, "pageNumber": 1, "totalItems": total},
	})
	return b
}

func TestSegmentList(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(listJSON([]api.Segment{{Name: "seg1"}, {Name: "seg2"}}, 2))
	}, "segment", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentGet(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON(api.Segment{Name: "internal_team", Type: "id_list"}))
	}, "segment", "get", "internal_team")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "internal_team" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestSegmentCreate(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		w.Write(entityJSON(api.Segment{Name: "new_seg"}))
	}, "segment", "create", "new_seg", "--type", "id_list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentDelete(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "delete", "old_seg")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestSegmentArchive(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"ok"}`))
	}, "segment", "archive", "old_seg")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["archived"] != true {
		t.Errorf("archived = %v", parsed["archived"])
	}
}

func TestSegmentIDsGet(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(entityJSON([]string{"user1", "user2"}))
	}, "segment", "ids", "get", "my_seg")

	if out == "" {
		t.Error("expected output")
	}
}

func TestSegmentIDsAdd(t *testing.T) {
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
	out, _ := runSegmentCmd(t, func(w http.ResponseWriter, r *http.Request) {
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
