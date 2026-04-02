package tag

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

func runTagCmd(t *testing.T, handler http.HandlerFunc, args ...string) (string, string) {
	t.Helper()
	shared.SetupMockServer(t, handler)

	root := &cobra.Command{Use: "test"}
	Register(root, func() *shared.GlobalFlags { return globals() })

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)

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

func TestTagList(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/tags" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("tags") != "" {
			t.Error("tags query param should not be sent for tag list")
		}
		w.Write(listJSON([]api.Tag{
			{ID: "t1", Name: "mobile", IsCore: true},
			{ID: "t2", Name: "web"},
		}, 2))
	}, "tag", "list")

	if out == "" {
		t.Error("expected output")
	}
}

func TestTagGet(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/console/v1/tags/tag-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write(entityJSON(api.Tag{ID: "tag-123", Name: "mobile", IsCore: true}))
	}, "tag", "get", "tag-123")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "mobile" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestTagCreate(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "mobile" {
			t.Errorf("name = %v", body["name"])
		}
		if body["description"] != "Mobile features" {
			t.Errorf("description = %v", body["description"])
		}
		if body["isCore"] != true {
			t.Errorf("isCore = %v", body["isCore"])
		}
		w.Write(entityJSON(api.Tag{ID: "t1", Name: "mobile", Description: "Mobile features", IsCore: true}))
	}, "tag", "create", "mobile", "--description", "Mobile features", "--is-core")

	if out == "" {
		t.Error("expected output")
	}
}

func TestTagUpdate(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/console/v1/tags/t1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "renamed" {
			t.Errorf("name = %v", body["name"])
		}
		w.Write(entityJSON(api.Tag{ID: "t1", Name: "renamed"}))
	}, "tag", "update", "t1", "--name", "renamed")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["name"] != "renamed" {
		t.Errorf("name = %v", parsed["name"])
	}
}

func TestTagDelete(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/console/v1/tags/t1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"message":"ok"}`))
	}, "tag", "delete", "t1")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	if parsed["status"] != "ok" {
		t.Errorf("status = %v", parsed["status"])
	}
}

func TestTagListWithSearch(t *testing.T) {
	out, _ := runTagCmd(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(listJSON([]api.Tag{
			{ID: "t1", Name: "mobile", Description: "Mobile"},
			{ID: "t2", Name: "web", Description: "Web"},
			{ID: "t3", Name: "mobile-ios", Description: "iOS"},
		}, 3))
	}, "tag", "list", "--search", "mobile")

	var parsed map[string]any
	json.Unmarshal([]byte(out), &parsed)
	data := parsed["data"].([]any)
	if len(data) != 2 {
		t.Errorf("expected 2 results, got %d", len(data))
	}
}
