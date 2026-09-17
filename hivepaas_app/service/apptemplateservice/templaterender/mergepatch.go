// Package templaterender turns a template and a user's choices into a
// specmodel.AppDoc. It is pure - no database, no network - so the linter in
// tools/apptemplate renders templates exactly the way HivePaaS does.
package templaterender

// MergePatch applies patch to target as an RFC 7396 JSON Merge Patch: maps
// merge recursively, a nil value deletes its key, and anything else - lists
// included - replaces wholesale. It returns a new tree and modifies neither
// argument.
func MergePatch(target, patch any) any {
	patchMap, isMap := patch.(map[string]any)
	if !isMap {
		return deepCopy(patch)
	}

	out := map[string]any{}
	if targetMap, ok := target.(map[string]any); ok {
		for key, value := range targetMap {
			out[key] = deepCopy(value)
		}
	}
	for key, value := range patchMap {
		if value == nil {
			delete(out, key)
			continue
		}
		out[key] = MergePatch(out[key], value)
	}
	return out
}

// deepCopy copies the maps and lists a decoded YAML tree is made of. Scalars
// are values already.
func deepCopy(node any) any {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			out[key] = deepCopy(value)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = deepCopy(value)
		}
		return out
	default:
		return v
	}
}
