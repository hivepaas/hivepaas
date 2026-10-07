package specmodel

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// maxNULDepth bounds the walk; a document is a tree, far shallower than this.
const maxNULDepth = 64

// FindNUL is where in v a string holds a NUL character - a key or a value - as
// a dotted path of the names the document is written with, and whether one does.
//
// Nothing that holds one can be written: a text column refuses it, and so does
// JSONB the \u0000 JSON writes for it. A YAML "\0" is one.
func FindNUL(v any) (at string, found bool) {
	return findNUL(reflect.ValueOf(v), "", 0)
}

func findNUL(v reflect.Value, at string, depth int) (string, bool) {
	if !v.IsValid() || depth > maxNULDepth {
		return "", false
	}
	switch v.Kind() { //nolint:exhaustive // the kinds that can hold a string
	case reflect.String:
		return at, strings.IndexByte(v.String(), 0) >= 0
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return "", false
		}
		return findNUL(v.Elem(), at, depth+1)
	case reflect.Struct:
		return findNULInStruct(v, at, depth)
	case reflect.Slice, reflect.Array:
		return findNULInList(v, at, depth)
	case reflect.Map:
		return findNULInMap(v, at, depth)
	}
	return "", false
}

func findNULInStruct(v reflect.Value, at string, depth int) (string, bool) {
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name, opts, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		if name == "-" {
			continue
		}
		next := joinPath(at, cmp.Or(name, field.Name))
		if strings.Contains(opts, "inline") {
			next = at
		}
		if found, ok := findNUL(v.Field(i), next, depth+1); ok {
			return found, true
		}
	}
	return "", false
}

func findNULInList(v reflect.Value, at string, depth int) (string, bool) {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		return "", false
	}
	for i := range v.Len() {
		if found, ok := findNUL(v.Index(i), fmt.Sprintf("%s[%d]", at, i), depth+1); ok {
			return found, true
		}
	}
	return "", false
}

// findNULInMap walks the keys in order, so that the first found is the same
// each time.
func findNULInMap(v reflect.Value, at string, depth int) (string, bool) {
	keys := v.MapKeys()
	slices.SortFunc(keys, func(a, b reflect.Value) int { return cmp.Compare(fmt.Sprint(a), fmt.Sprint(b)) })
	for _, key := range keys {
		name := fmt.Sprint(key)
		next := joinPath(at, name)
		if strings.IndexByte(name, 0) >= 0 {
			return next, true
		}
		if found, ok := findNUL(v.MapIndex(key), next, depth+1); ok {
			return found, true
		}
	}
	return "", false
}

func joinPath(at, name string) string {
	if at == "" {
		return name
	}
	return at + "." + name
}
