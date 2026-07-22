package shared

import (
	"os"

	"github.com/shhac/agent-statsig/internal/api"
	"github.com/shhac/agent-statsig/internal/output"
)

// WriteResource writes a single resource in the resolved format (NDJSON, JSON,
// or YAML); the format defaults to JSON when the flag is empty or unparseable.
func WriteResource(data any, format string) {
	output.Print(data, output.ResolveFormat(format), true)
}

func WritePaginatedList(items []any, pagination *api.PaginationInfo, format string) {
	f := output.ResolveFormat(format)
	if f == output.FormatNDJSON {
		w := output.NewNDJSONWriter(os.Stdout)
		for _, item := range items {
			_ = w.WriteItem(item)
		}
		if pagination != nil {
			_ = w.WritePagination(&output.Pagination{
				HasMore:    pagination.HasMore(),
				TotalItems: pagination.TotalItems,
				Page:       pagination.PageNumber,
			})
		}
		return
	}
	result := map[string]any{"data": items}
	if pagination != nil {
		result["pagination"] = map[string]any{
			"hasMore":    pagination.HasMore(),
			"totalItems": pagination.TotalItems,
			"page":       pagination.PageNumber,
		}
	}
	output.Print(result, f, true)
}
