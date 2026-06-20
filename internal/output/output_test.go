package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
	out "github.com/shhac/lib-agent-output"
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

// TestYAMLEncoderRegistered pins that `--format yaml` actually works: the
// blank-imported lib-agent-cli/yaml package registers a YAML encoder with
// lib-agent-output, so out.Print(FormatYAML) emits valid, 2-space-indented YAML
// rather than erroring on an unregistered format. Exercising out.Print directly
// (writer-injectable) hits the same encoder the CLI's Print routes through.
func TestYAMLEncoderRegistered(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"name": "gate", "isEnabled": true}
	if err := out.Print(&buf, data, FormatYAML, out.PruneNils); err != nil {
		t.Fatalf("out.Print(FormatYAML) error: %v (encoder not registered?)", err)
	}

	got := buf.String()
	if !strings.Contains(got, "name: gate") || !strings.Contains(got, "isEnabled: true") {
		t.Errorf("YAML output missing expected fields:\n%s", got)
	}
}

// TestYAMLNormalizesWholeNumbers pins the latent-bug fix the shared encoder
// brings: JSON decoding produces float64 for every number, and yaml.v3 renders
// large whole floats in scientific notation (e.g. "1.5e+06"). The shared
// encoder normalizes whole-valued floats to integers, so a large id/count
// renders as a plain integer.
func TestYAMLNormalizesWholeNumbers(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{"count": float64(1500000)}
	if err := out.Print(&buf, data, FormatYAML, nil); err != nil {
		t.Fatalf("out.Print(FormatYAML) error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "count: 1500000") {
		t.Errorf("whole-number not normalized to integer:\n%s", got)
	}
	if strings.Contains(got, "e+") {
		t.Errorf("YAML used scientific notation for a whole number:\n%s", got)
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
