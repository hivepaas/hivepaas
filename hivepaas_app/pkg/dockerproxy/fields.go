package dockerproxy

import (
	"encoding/json"
	"maps"
	"slices"
)

// neutralValues are what a client sends for "not set" where zero would mean
// something. The docker CLI sends MemorySwappiness -1.
var neutralValues = map[string][]string{
	"MemorySwappiness": {"-1"},
}

// isZero reports whether v is what a client sends for a field it did not set.
// Go clients marshal whole structs, so most of a HostConfig arrives as null, "",
// 0, false, an empty list, or an object holding only those.
func isZero(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case bool:
		return !value
	case string:
		return value == ""
	case int64:
		return value == 0
	case json.Number:
		f, err := value.Float64()
		return err == nil && f == 0
	case []any:
		return len(value) == 0
	case map[string]any:
		for _, inner := range value {
			if !isZero(inner) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isNeutral(field string, v any) bool {
	n, ok := v.(json.Number)
	return ok && slices.Contains(neutralValues[field], n.String())
}

// checkFields refuses any field of obj outside allowed that carries a value.
// A field in nullOnly must be absent or null, because an empty value there
// means something: an empty MaskedPaths unmasks /proc.
func checkFields(where string, obj map[string]any, allowed, nullOnly []string) error {
	for _, field := range slices.Sorted(maps.Keys(obj)) {
		value := obj[field]
		switch {
		case slices.Contains(nullOnly, field):
			if value != nil {
				return refusef("%s.%s must not be set", where, field)
			}
		case slices.Contains(allowed, field), isZero(value), isNeutral(field, value):
		default:
			return refusef("%s.%s is not allowed", where, field)
		}
	}
	return nil
}
