package shared

import (
	"strings"
)

// ToAnySlice converts a typed slice to []any.
func ToAnySlice[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// FilterBySearch filters a slice by substring match on name/description fields.
func FilterBySearch[T any](items []T, search string, getName func(T) string, getDesc func(T) string) []T {
	search = strings.ToLower(search)
	var filtered []T
	for _, item := range items {
		if strings.Contains(strings.ToLower(getName(item)), search) ||
			strings.Contains(strings.ToLower(getDesc(item)), search) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func SliceRemove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

func ToStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	default:
		return nil
	}
}

func MapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
