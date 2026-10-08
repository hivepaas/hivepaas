package main

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// fixStats counts what fixSpec changed, for the summary a run prints.
type fixStats struct {
	schemas    int
	optional   int
	nullable   int
	anyValue   int
	unresolved []string
}

func (s fixStats) String() string {
	return fmt.Sprintf("%d schemas from Go types: %d fields optional, %d nullable, %d any value; "+
		"%d schemas with no Go type found", s.schemas, s.optional, s.nullable, s.anyValue, len(s.unresolved))
}

// fixSpec corrects what swag cannot say about the API's types, from the Go
// types the schemas are named after:
//
//   - a field with omitempty is not required: it is left out when empty. Swag's
//     --requiredByDefault makes every field required, omitempty or not.
//   - a field that can be null - a pointer, a slice or a map without omitempty,
//     which encoding/json writes as null when nil - is nullable.
//   - a field of type any, or json.RawMessage, is any JSON value. Swag writes
//     an object, and the converter adds the type to an empty schema.
//
// A field given its schema by hand, with a swaggertype tag, is only made
// optional.
func fixSpec(doc *object, src *goSource) (fixStats, error) {
	var stats fixStats
	components := doc.getObject("components")
	if components == nil {
		return stats, fmt.Errorf("the document has no components")
	}
	schemas := components.getObject("schemas")
	if schemas == nil {
		return stats, fmt.Errorf("the document has no schemas")
	}
	for _, name := range schemas.keys {
		schema := schemas.getObject(name)
		decl := src.schemaType(name)
		if schema == nil || decl == nil {
			if schema != nil && schema.getObject("properties") != nil {
				stats.unresolved = append(stats.unresolved, name)
			}
			continue
		}
		stats.schemas++
		fixObjectSchema(schema, src.jsonFields(decl), src, &stats)
	}
	sort.Strings(stats.unresolved)
	return stats, nil
}

// schemaType is the Go struct a schema is named after: package.Type, which is
// how swag names them once swag.sh has shortened the package paths. A name two
// packages could give is left alone.
func (s *goSource) schemaType(name string) *typeDecl {
	pkgName, typeName, found := strings.Cut(name, ".")
	if !found || strings.Contains(typeName, ".") {
		return nil
	}
	var match *typeDecl
	for _, pkg := range s.byName[pkgName] {
		if decl := pkg.types[typeName]; decl != nil {
			if match != nil {
				return nil
			}
			match = decl
		}
	}
	if match == nil {
		return nil
	}
	if st, _ := s.structOf(match); st == nil {
		return nil
	}
	return match
}

func fixObjectSchema(schema *object, fields []jsonField, src *goSource, stats *fixStats) {
	properties := schema.getObject("properties")
	if properties == nil {
		return
	}
	required, _ := schema.get("required").([]any)
	for _, field := range fields {
		property := properties.getObject(field.name)
		if property == nil {
			continue
		}
		if field.omitEmpty && slices.Contains(required, any(field.name)) {
			required = slices.DeleteFunc(required, func(v any) bool { return v == field.name })
			stats.optional++
		}
		if field.swaggerType {
			continue
		}
		fixed := fixValueSchema(property, field.ref, src, stats)
		if !field.omitEmpty && src.shape(field.ref).kind == kindNullable {
			if nullable, changed := makeNullable(fixed); changed {
				fixed = nullable
				stats.nullable++
			}
		}
		properties.set(field.name, fixed)
	}
	if len(required) == 0 {
		schema.remove("required")
	} else {
		schema.set("required", required)
	}
}

// fixValueSchema makes the schema of a value of type any, wherever it is - the
// field itself, a slice's items, a map's values - the schema of any value.
func fixValueSchema(schema *object, ref typeRef, src *goSource, stats *fixStats) *object {
	for range maxTypeDepth {
		shape := src.shape(ref)
		switch {
		case shape.kind == kindAny:
			stats.anyValue++
			return anyValueSchema(schema)
		case shape.kind != kindNullable || shape.elem == nil:
			return schema
		case shape.isSlice:
			if items := schema.getObject("items"); items != nil {
				schema.set("items", fixValueSchema(items, *shape.elem, src, stats))
			}
			return schema
		case shape.isMap:
			if values := schema.getObject("additionalProperties"); values != nil {
				schema.set("additionalProperties", fixValueSchema(values, *shape.elem, src, stats))
			}
			return schema
		default: // a pointer: what it points at decides
			ref = *shape.elem
		}
	}
	return schema
}

// anyValueSchema keeps a schema's description and nothing that would limit
// the value: no type.
func anyValueSchema(schema *object) *object {
	out := newObject()
	if description, ok := schema.get("description").(string); ok {
		out.set("description", description)
	}
	return out
}

// makeNullable lets a schema be null. A reference cannot carry anything beside
// it in OpenAPI 3.0, so it goes inside an allOf, as swag does for a reference
// with a description. A schema without a type is any value, null included.
func makeNullable(schema *object) (*object, bool) {
	if nullable, _ := schema.get("nullable").(bool); nullable {
		return schema, false
	}
	if ref, ok := schema.get("$ref").(string); ok {
		out := newObject()
		item := newObject()
		item.set("$ref", ref)
		out.set("allOf", []any{item})
		if description, ok := schema.get("description").(string); ok {
			out.set("description", description)
		}
		out.set("nullable", true)
		return out, true
	}
	if schema.get("type") == nil && schema.get("allOf") == nil {
		return schema, false
	}
	schema.set("nullable", true)
	return schema, true
}
