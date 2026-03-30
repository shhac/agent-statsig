package output

import (
	"bytes"
	"encoding/json"
	"testing"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
		err   bool
	}{
		{"json", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"jsonl", FormatNDJSON, false},
		{"ndjson", FormatNDJSON, false},
		{"csv", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("ParseFormat(%q) should error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseFormat(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveFormat(t *testing.T) {
	if got := ResolveFormat("yaml"); got != FormatYAML {
		t.Errorf("ResolveFormat('yaml') = %q, want yaml", got)
	}
	if got := ResolveFormat(""); got != FormatJSON {
		t.Errorf("ResolveFormat('') = %q, want json", got)
	}
	if got := ResolveFormat("garbage"); got != FormatJSON {
		t.Errorf("ResolveFormat('garbage') = %q, want json (fallback)", got)
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	err := agenterrors.New("test error", agenterrors.FixableByAgent).WithHint("try again")
	WriteError(&buf, err)

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if parsed["error"] != "test error" {
		t.Errorf("error = %v", parsed["error"])
	}
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v", parsed["fixable_by"])
	}
	if parsed["hint"] != "try again" {
		t.Errorf("hint = %v", parsed["hint"])
	}
}

func TestWriteErrorPlain(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, agenterrors.New("plain", agenterrors.FixableByHuman))

	var parsed map[string]any
	json.Unmarshal(buf.Bytes(), &parsed)
	if _, ok := parsed["hint"]; ok {
		t.Error("hint should be absent when not set")
	}
}

func TestWriteErrorNonAPIError(t *testing.T) {
	var buf bytes.Buffer
	WriteError(&buf, &simpleErr{msg: "boom"})

	var parsed map[string]any
	json.Unmarshal(buf.Bytes(), &parsed)
	if parsed["error"] != "boom" {
		t.Errorf("error = %v", parsed["error"])
	}
	if parsed["fixable_by"] != "agent" {
		t.Errorf("fixable_by = %v, want agent (default)", parsed["fixable_by"])
	}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

func TestNDJSONWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewNDJSONWriter(&buf)

	w.WriteItem(map[string]any{"name": "gate1"})
	w.WriteItem(map[string]any{"name": "gate2"})
	w.WritePagination(&Pagination{HasMore: true, TotalItems: 100, Page: 1})

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}

	var item1 map[string]any
	json.Unmarshal(lines[0], &item1)
	if item1["name"] != "gate1" {
		t.Errorf("line 1 name = %v", item1["name"])
	}

	var pag map[string]any
	json.Unmarshal(lines[2], &pag)
	inner := pag["@pagination"].(map[string]any)
	if inner["hasMore"] != true {
		t.Error("pagination hasMore should be true")
	}
}

func TestPruneNulls(t *testing.T) {
	input := map[string]any{
		"name":  "test",
		"value": nil,
		"nested": map[string]any{
			"a": 1,
			"b": nil,
		},
	}
	result := pruneNulls(input).(map[string]any)
	if _, ok := result["value"]; ok {
		t.Error("null value should be pruned")
	}
	nested := result["nested"].(map[string]any)
	if _, ok := nested["b"]; ok {
		t.Error("nested null should be pruned")
	}
	if nested["a"] != 1 {
		t.Errorf("nested.a = %v", nested["a"])
	}
}
