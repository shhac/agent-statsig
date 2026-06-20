// Package output re-exports the shared output contract from lib-agent-output,
// keeping the internal/output import path while the wire mechanism (format
// parsing, JSON/YAML encoding, error rendering) lives in one place. What stays
// local is agent-statsig policy: the format-routed Print, the one-arg
// ResolveFormat that defaults to JSON, and the Statsig-shaped pagination
// trailer. (Migration shim.)
package output

import (
	"io"
	"os"

	out "github.com/shhac/lib-agent-output"

	// Registers the canonical YAML encoder (2-space indent, whole-float-to-int
	// normalization) with lib-agent-output for its side effect, so this CLI gets
	// `--format yaml` without carrying a local encoder.
	_ "github.com/shhac/lib-agent-cli/yaml"
)

// Format and its values come from the shared contract; ParseFormat is therefore
// the family's lenient parser (accepts "ndjson"/"yml", case-insensitive).
type Format = out.Format

const (
	FormatJSON   = out.FormatJSON
	FormatYAML   = out.FormatYAML
	FormatNDJSON = out.FormatNDJSON
)

// ParseFormat is the shared lenient parser. WriteError renders the structured
// {error, fixable_by, hint} line via the shared encoder (HTML escaping off).
var (
	ParseFormat = out.ParseFormat
	WriteError  = out.WriteError
)

// ResolveFormat keeps agent-statsig's one-arg, always-default-JSON behavior:
// an empty or unparseable flag resolves to JSON rather than surfacing an error.
func ResolveFormat(flagFormat string) Format {
	if flagFormat == "" {
		return FormatJSON
	}
	f, err := ParseFormat(flagFormat)
	if err != nil {
		return FormatJSON
	}
	return f
}

// Print writes data to stdout in the given format. When prune is true, null
// values are removed; when false the data is still round-tripped (via
// identityPruner) so raw JSON is decoded — re-indented for JSON, and rendered as
// real YAML mappings rather than a quoted blob for YAML. NDJSON streams a single
// line. YAML is delegated to the encoder registered by the blank-imported
// lib-agent-cli/yaml package.
func Print(data any, format Format, prune bool) {
	pruner := identityPruner
	if prune {
		pruner = out.PruneNils
	}
	_ = out.Print(os.Stdout, data, format, pruner)
}

// PrintJSON pretty-prints data to stdout. When prune is true, null values are
// removed; when false the data is still round-tripped so raw JSON is re-indented
// (matching the pre-migration behavior).
func PrintJSON(data any, prune bool) {
	Print(data, FormatJSON, prune)
}

// identityPruner forces out.Print's normalize round-trip without dropping any
// fields, so a json.RawMessage gets decoded and re-indented like the old
// PrintJSON did.
func identityPruner(v any) any { return v }

// NDJSONWriter writes one JSON object per line to the given writer.
type NDJSONWriter struct {
	w *out.NDJSONWriter
}

func NewNDJSONWriter(w io.Writer) *NDJSONWriter {
	return &NDJSONWriter{w: out.NewNDJSONWriter(w)}
}

func (n *NDJSONWriter) WriteItem(item any) error {
	return n.w.WriteItem(item)
}

func (n *NDJSONWriter) WritePagination(p *Pagination) error {
	return n.w.WriteMetaLine("@pagination", p)
}

// Pagination is Statsig-shaped (page/total counters, not an opaque cursor), so
// it stays local rather than using out.Pagination.
type Pagination struct {
	HasMore    bool `json:"hasMore"`
	TotalItems int  `json:"totalItems"`
	Page       int  `json:"page"`
}
