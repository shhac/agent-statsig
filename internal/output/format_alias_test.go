package output

import "testing"

// TestParseFormatLenient pins the now-shared (intentionally lenient) parser:
// case-insensitive, with yml/ndjson aliases accepted, and unrelated formats
// still rejected (fixable_by: agent). This guards the contract after the
// migration routed ParseFormat through lib-agent-output.
func TestParseFormatLenient(t *testing.T) {
	accepted := map[string]Format{
		"json":   FormatJSON,
		"JSON":   FormatJSON,
		"yaml":   FormatYAML,
		"yml":    FormatYAML,
		"YAML":   FormatYAML,
		"jsonl":  FormatNDJSON,
		"ndjson": FormatNDJSON,
	}
	for in, want := range accepted {
		got, err := ParseFormat(in)
		if err != nil {
			t.Errorf("ParseFormat(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseFormat(%q) = %q, want %q", in, got, want)
		}
	}

	for _, in := range []string{"toml", "csv", "xml", ""} {
		if _, err := ParseFormat(in); err == nil {
			t.Errorf("ParseFormat(%q) should be rejected", in)
		}
	}
}
