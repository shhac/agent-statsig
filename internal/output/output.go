package output

import (
	"encoding/json"
	"io"
	"os"

	agenterrors "github.com/shhac/agent-statsig/internal/errors"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatNDJSON Format = "jsonl"
)

func ParseFormat(s string) (Format, error) {
	switch s {
	case "json":
		return FormatJSON, nil
	case "yaml":
		return FormatYAML, nil
	case "jsonl", "ndjson":
		return FormatNDJSON, nil
	default:
		return "", agenterrors.Newf(agenterrors.FixableByAgent, "unknown format %q, expected: json, yaml, jsonl", s)
	}
}

func ResolveFormat(flagFormat string) Format {
	if flagFormat != "" {
		f, err := ParseFormat(flagFormat)
		if err != nil {
			return FormatJSON
		}
		return f
	}
	return FormatJSON
}

// PrintJSON pretty-prints data to stdout. When prune is true, null values are removed.
func PrintJSON(data any, prune bool) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	if prune {
		var m any
		if err := json.Unmarshal(b, &m); err == nil {
			m = pruneNulls(m)
			b, _ = json.Marshal(m)
		}
	}
	var indented any
	if err := json.Unmarshal(b, &indented); err == nil {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(indented)
	}
}

// WriteError writes a structured JSON error to the given writer.
func WriteError(w io.Writer, err error) {
	var aerr *agenterrors.APIError
	if !agenterrors.As(err, &aerr) {
		aerr = agenterrors.Wrap(err, agenterrors.FixableByAgent)
	}
	payload := map[string]any{
		"error":      aerr.Message,
		"fixable_by": string(aerr.FixableBy),
	}
	if aerr.Hint != "" {
		payload["hint"] = aerr.Hint
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// NDJSONWriter writes one JSON object per line to the given writer.
type NDJSONWriter struct {
	enc *json.Encoder
}

func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &NDJSONWriter{enc: enc}
}

func (n *NDJSONWriter) WriteItem(item any) error {
	return n.enc.Encode(item)
}

func (n *NDJSONWriter) WritePagination(p *Pagination) error {
	return n.enc.Encode(map[string]any{"@pagination": p})
}

type Pagination struct {
	HasMore    bool `json:"hasMore"`
	TotalItems int  `json:"totalItems"`
	Page       int  `json:"page"`
}

func pruneNulls(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v := range val {
			if v == nil {
				continue
			}
			out[k] = pruneNulls(v)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = pruneNulls(v)
		}
		return out
	default:
		return v
	}
}
