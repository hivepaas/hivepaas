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
