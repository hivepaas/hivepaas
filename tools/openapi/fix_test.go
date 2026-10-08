package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testModule = "example.com/m"

func loadTestSource(t *testing.T) *goSource {
	t.Helper()
	src, err := loadGoSource("testdata/module", testModule, "hivepaas_app")
	assert.NoError(t, err)
	return src
}

// The schema swag writes for demodto.ItemResp, as the converter leaves it:
// every tagged field required, an any an object.
const itemSchemaJSON = `{
  "components": {"schemas": {
    "demodto.Meta": {"properties": {"total": {"type": "integer"}}, "type": "object"},
    "demodto.ItemResp": {
      "properties": {
        "default": {"description": "Default is any value.", "type": "object"},
        "id": {"type": "string"},
        "labels": {"additionalProperties": {"type": "string"}, "type": "object"},
        "meta": {"$ref": "#/components/schemas/demodto.Meta"},
        "name": {"type": "string"},
        "nick": {"type": "string"},
        "note": {"type": "string"},
        "params": {"additionalProperties": {"type": "object"}, "type": "object"},
        "raw": {"items": {"type": "integer"}, "type": "array"},
        "size": {"type": "string"},
        "tags": {"items": {"type": "string"}, "type": "array"},
        "typed": {"type": "string"}
      },
      "required": ["default", "id", "labels", "meta", "name", "nick", "note", "params", "raw", "size", "tags", "typed"],
      "type": "object"
    },
    "inline_request": {"properties": {"file": {"type": "string"}}, "type": "object"}
  }}
}`

func TestFixSpecSaysWhatTheGoTypesMean(t *testing.T) {
	value, err := decodeDocument([]byte(itemSchemaJSON))
	assert.NoError(t, err)
	doc := value.(*object)

	stats, err := fixSpec(doc, loadTestSource(t))

	assert.NoError(t, err)
	assert.Equal(t, []string{"inline_request"}, stats.unresolved)
	got := string(encodeDocument(doc.getObject("components").getObject("schemas").getObject("demodto.ItemResp")))
	assert.Equal(t, `{
  "properties" : {
    "default" : {
      "description" : "Default is any value."
    },
    "id" : {
      "type" : "string"
    },
    "labels" : {
      "additionalProperties" : {
        "type" : "string"
      },
      "nullable" : true,
      "type" : "object"
    },
    "meta" : {
      "allOf" : [ {
        "$ref" : "#/components/schemas/demodto.Meta"
      } ],
      "nullable" : true
    },
    "name" : {
      "type" : "string"
    },
    "nick" : {
      "type" : "string"
    },
    "note" : {
      "nullable" : true,
      "type" : "string"
    },
    "params" : {
      "additionalProperties" : { },
      "nullable" : true,
      "type" : "object"
    },
    "raw" : { },
    "size" : {
      "type" : "string"
    },
    "tags" : {
      "items" : {
        "type" : "string"
      },
      "nullable" : true,
      "type" : "array"
    },
    "typed" : {
      "type" : "string"
    }
  },
  "required" : [ "id", "labels", "meta", "name", "note", "params", "raw", "size", "tags", "typed" ],
  "type" : "object"
}`, got, "omitempty is optional, a pointer, map or slice nullable, any any value; a type that writes "+
		"itself and one given its schema by hand are left alone")
}

// A Go constant that aliases another is not a value of its own: the enum lists
// each value once, and its names stay lined up with the values.
func TestFixSpecListsAnEnumsValuesOnce(t *testing.T) {
	value, err := decodeDocument([]byte(`{"components": {"schemas": {
		"base.KeyType": {"enum": ["ec", "rsa", "ec"], "type": "string",
			"x-enum-varnames": ["KeyTypeEC", "KeyTypeRSA", "KeyTypeDefault"]},
		"base.Mode": {"enum": ["a", "a"], "type": "string", "x-enum-varnames": ["ModeA"]},
		"base.Fine": {"enum": [1, 2], "type": "integer"}
	}}}`))
	assert.NoError(t, err)
	doc := value.(*object)

	stats, err := fixSpec(doc, loadTestSource(t))

	assert.NoError(t, err)
	assert.Equal(t, 2, stats.enums)
	schemas := doc.getObject("components").getObject("schemas")
	assert.Equal(t, `{
  "enum" : [ "ec", "rsa" ],
  "type" : "string",
  "x-enum-varnames" : [ "KeyTypeEC", "KeyTypeRSA" ]
}`, string(encodeDocument(schemas.getObject("base.KeyType"))))
	assert.Equal(t, `{
  "enum" : [ "a" ],
  "type" : "string",
  "x-enum-varnames" : [ "ModeA" ]
}`, string(encodeDocument(schemas.getObject("base.Mode"))), "names that never lined up are left as they are")
}

// A body swag gave no media type is the JSON the handlers read; a form, a body
// without a schema of the API's own, and an operation without a body stay.
func TestFixSpecMakesRequestBodiesJSON(t *testing.T) {
	value, err := decodeDocument([]byte(`{"paths": {"/items": {
		"post": {"requestBody": {"content": {"*/*": {"schema": {"$ref": "#/components/schemas/demo.Item"}}}}},
		"put": {"requestBody": {"content": {"multipart/form-data": {"schema": {"type": "object"}}}}},
		"patch": {"requestBody": {"content": {"*/*": {"schema": {"type": "string"}}}}},
		"delete": {}
	}}, "components": {"schemas": {}}}`))
	assert.NoError(t, err)
	doc := value.(*object)

	stats, err := fixSpec(doc, loadTestSource(t))

	assert.NoError(t, err)
	assert.Equal(t, 1, stats.jsonBodies)
	item := doc.getObject("paths").getObject("/items")
	assert.Equal(t, []string{"application/json"}, item.getObject("post").getObject("requestBody").getObject("content").keys)
	assert.Equal(t, []string{"multipart/form-data"}, item.getObject("put").getObject("requestBody").getObject("content").keys)
	assert.Equal(t, []string{"*/*"}, item.getObject("patch").getObject("requestBody").getObject("content").keys)
}
