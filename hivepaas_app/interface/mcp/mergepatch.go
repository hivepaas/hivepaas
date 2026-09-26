package mcp

import (
	"reflect"
	"slices"
	"strings"
)

// mergePatch applies a JSON merge patch (RFC 7396) to a document, both decoded
// from JSON: a patch's object merges into the document's field by field, null
// removes a field, and anything else - a list included - replaces the value.
func mergePatch(doc, patch any) any {
	patchObj, ok := patch.(map[string]any)
	if !ok {
		return patch
	}
	docObj, ok := doc.(map[string]any)
	if !ok {
		docObj = map[string]any{}
	}
	out := make(map[string]any, len(docObj))
	for k, v := range docObj {
		out[k] = v
	}
	for k, v := range patchObj {
		if v == nil {
			delete(out, k)
			continue
		}
		out[k] = mergePatch(out[k], v)
	}
	return out
}

// fieldChange is one value a change makes different.
type fieldChange struct {
	Path   string `json:"path"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

// diffFields lists what differs between two documents decoded from JSON, a
// field of an object at a time; a list that differs is one change, whole.
func diffFields(before, after any) []fieldChange {
	var out []fieldChange
	diffAt("", before, after, &out)
	return out
}

func diffAt(path string, before, after any, out *[]fieldChange) {
	beforeObj, beforeIsObj := before.(map[string]any)
	afterObj, afterIsObj := after.(map[string]any)
	if !beforeIsObj || !afterIsObj {
		if !reflect.DeepEqual(before, after) {
			*out = append(*out, fieldChange{Path: path, Before: before, After: after})
		}
		return
	}
	keys := make([]string, 0, len(beforeObj)+len(afterObj))
	for k := range beforeObj {
		keys = append(keys, k)
	}
	for k := range afterObj {
		if _, ok := beforeObj[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	for _, k := range keys {
		diffAt(joinPath(path, k), beforeObj[k], afterObj[k], out)
	}
}

func joinPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// droppedFields are the top-level fields of a patch that a document does not
// have once it has been through the endpoint's request type: fields the
// endpoint does not take, which a change of them would silently lose.
func droppedFields(patch map[string]any, taken map[string]any) []string {
	var dropped []string
	for k, v := range patch {
		if v == nil {
			continue
		}
		if _, ok := taken[k]; !ok {
			dropped = append(dropped, k)
		}
	}
	slices.Sort(dropped)
	return dropped
}

// describePaths is paths in a sentence.
func describePaths(paths []string) string {
	return strings.Join(paths, ", ")
}
