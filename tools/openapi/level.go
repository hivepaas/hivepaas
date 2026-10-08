package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	apiLevelFile   = "docs/openapi/api-level.json"
	apiLevelGoFile = "hivepaas_app/base/api_level_gen.go"
	apiLevelKey    = "x-api-level"
	// maxNamedOperations is how many changed operations a run names.
	maxNamedOperations = 5
)

// writeMethods are the operations whose requests the API level follows: a client
// that reads what it does not know loses nothing, one that writes it back does.
var writeMethods = []string{"post", "put", "patch", "delete"}

// docKeys are the parts of a schema that say nothing about what a request
// carries: a change to them is not a change to the request.
var docKeys = map[string]bool{
	"description": true, "summary": true, "title": true, "example": true, "examples": true,
	"deprecated": true, "x-enum-varnames": true, "x-enum-descriptions": true, "x-enum-comments": true,
}

// apiLevel is docs/openapi/api-level.json: the level, and the fingerprint of the
// request of every write operation at that level.
type apiLevel struct {
	Level      int               `json:"level"`
	Operations map[string]string `json:"operations"`
}

// runLevel keeps the API level: the number a HivePaaS CLI sends to say which API
// it was built for, and below which the server refuses its writes.
//
// It is raised when the request of a write operation that already existed
// changes - a field added to its body, a parameter added - because a CLI built
// before that change would write the object back without it. An operation added,
// or a change to what an operation answers, does not raise it: an older CLI
// neither calls the one nor loses anything to the other.
//
// It writes the level to docs/openapi/api-level.json with the fingerprints it
// compares the next run against, to the spec's info as x-api-level, and to
// base.APILevel for the server. Run on an unchanged spec, it changes nothing.
func runLevel(args []string, out io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, _, err := moduleRoot(wd)
	if err != nil {
		return fmt.Errorf("no go.mod above %s", wd)
	}
	file := filepath.Join(root, defaultSpec)
	if len(args) > 0 {
		file = args[0]
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	value, err := decodeDocument(data)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	doc, ok := value.(*object)
	if !ok {
		return fmt.Errorf("%s: not a JSON object", file)
	}

	ops, err := operationFingerprints(doc)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	prev, err := readAPILevel(filepath.Join(root, apiLevelFile))
	if err != nil {
		return err
	}
	next, changed := nextAPILevel(prev, ops)

	levelJSON, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(root, apiLevelFile), append(levelJSON, '\n'), 0o644); err != nil { //nolint:gosec
		return err
	}
	if err = os.WriteFile(filepath.Join(root, apiLevelGoFile), apiLevelGo(next.Level), 0o644); err != nil { //nolint:gosec
		return err
	}
	if info := doc.getObject("info"); info != nil {
		info.set(apiLevelKey, json.Number(fmt.Sprint(next.Level)))
	}
	if err = os.WriteFile(file, encodeDocument(doc), 0o644); err != nil { //nolint:gosec
		return err
	}

	if len(changed) > 0 {
		named := changed
		if len(named) > maxNamedOperations {
			named = append(slices.Clone(named[:maxNamedOperations]),
				fmt.Sprintf("%d more", len(changed)-maxNamedOperations))
		}
		fmt.Fprintf(out, "API level %d, raised from %d: the request of %s changed\n",
			next.Level, prev.Level, strings.Join(named, ", "))
		return nil
	}
	fmt.Fprintf(out, "API level %d\n", next.Level)
	return nil
}

func readAPILevel(path string) (apiLevel, error) {
	level := apiLevel{Operations: map[string]string{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return level, nil
	}
	if err != nil {
		return level, err
	}
	if err = json.Unmarshal(data, &level); err != nil {
		return level, fmt.Errorf("%s: %w", path, err)
	}
	return level, nil
}

// nextAPILevel is the level for the operations as they are now, and the
// operations whose request changed since prev.
func nextAPILevel(prev apiLevel, ops map[string]string) (apiLevel, []string) {
	var changed []string
	for op, sum := range ops {
		if old, found := prev.Operations[op]; found && old != sum {
			changed = append(changed, op)
		}
	}
	sort.Strings(changed)
	level := prev.Level
	switch {
	case level == 0:
		level = 1
	case len(changed) > 0:
		level++
	}
	return apiLevel{Level: level, Operations: ops}, changed
}

// operationFingerprints are the fingerprints of the requests of the write
// operations: their parameters and body, references resolved, without their
// documentation.
func operationFingerprints(doc *object) (map[string]string, error) {
	paths := doc.getObject("paths")
	if paths == nil {
		return nil, errors.New("the document has no paths")
	}
	ops := map[string]string{}
	for _, path := range paths.keys {
		item := paths.getObject(path)
		if item == nil {
			continue
		}
		for _, method := range writeMethods {
			op := item.getObject(method)
			if op == nil {
				continue
			}
			request := map[string]any{
				"parameters":  canonicalParameters(doc, item.get("parameters"), op.get("parameters")),
				"requestBody": canonical(doc, op.get("requestBody"), nil),
			}
			data, err := json.Marshal(request)
			if err != nil {
				return nil, err
			}
			sum := sha256.Sum256(data)
			ops[strings.ToUpper(method)+" "+path] = hex.EncodeToString(sum[:])
		}
	}
	return ops, nil
}

// canonicalParameters are a path's and an operation's parameters, in an order
// that does not depend on the order they were documented in.
func canonicalParameters(doc *object, lists ...any) []any {
	var params []any
	for _, list := range lists {
		if items, ok := list.([]any); ok {
			for _, item := range items {
				params = append(params, canonical(doc, item, nil))
			}
		}
	}
	sort.Slice(params, func(i, j int) bool { return paramKey(params[i]) < paramKey(params[j]) })
	return params
}

func paramKey(param any) string {
	m, _ := param.(map[string]any)
	return fmt.Sprintf("%v %v", m["in"], m["name"])
}

// canonical is a schema as plain values with its references resolved - a
// reference already being resolved on the way stays one, so a recursive schema
// ends - without what only documents it, with required names and enum values in
// a fixed order.
func canonical(doc *object, node any, resolving []string) any {
	switch n := node.(type) {
	case *object:
		if ref, ok := n.get("$ref").(string); ok {
			if slices.Contains(resolving, ref) {
				return map[string]any{"$ref": ref}
			}
			target := resolveRef(doc, ref)
			if target == nil {
				return map[string]any{"$ref": ref}
			}
			return canonical(doc, target, append(slices.Clone(resolving), ref))
		}
		out := map[string]any{}
		for _, key := range n.keys {
			if docKeys[key] {
				continue
			}
			value := canonical(doc, n.values[key], resolving)
			if key == "required" || key == "enum" {
				value = sortedValues(value)
			}
			out[key] = value
		}
		return out
	case []any:
		out := make([]any, 0, len(n))
		for _, item := range n {
			out = append(out, canonical(doc, item, resolving))
		}
		return out
	default:
		return n
	}
}

func sortedValues(value any) any {
	items, ok := value.([]any)
	if !ok {
		return value
	}
	sorted := slices.Clone(items)
	sort.Slice(sorted, func(i, j int) bool { return fmt.Sprint(sorted[i]) < fmt.Sprint(sorted[j]) })
	return sorted
}

// resolveRef finds a local reference, #/components/schemas/name, in the document.
func resolveRef(doc *object, ref string) any {
	pointer, found := strings.CutPrefix(ref, "#/")
	if !found {
		return nil
	}
	var node any = doc
	for _, part := range strings.Split(pointer, "/") {
		obj, ok := node.(*object)
		if !ok {
			return nil
		}
		node = obj.get(strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~"))
	}
	return node
}

func apiLevelGo(level int) []byte {
	return []byte(fmt.Sprintf(`// Code generated by tools/openapi level from %s. DO NOT EDIT.

package base

// APILevel is the level of the API this server answers. It is raised whenever
// the request of a write operation that already existed changes, and a HivePaaS
// CLI built for a lower level may not write: it would send back objects without
// the fields it does not know. See tools/openapi/level.go.
const APILevel = %d
`, apiLevelFile, level))
}
