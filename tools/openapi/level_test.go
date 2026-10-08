package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// specWith is a spec with one write operation, whose body is item, and one read.
func specWith(t *testing.T, item string, params string) *object {
	t.Helper()
	value, err := decodeDocument([]byte(`{
  "paths": {
    "/items/{itemID}": {
      "get": {"parameters": [{"in": "query", "name": "fresh", "schema": {"type": "boolean"}}]},
      "put": {
        "parameters": ` + params + `,
        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/demo.Item"}}}}
      }
    }
  },
  "components": {"schemas": {"demo.Item": ` + item + `}}
}`))
	assert.NoError(t, err)
	return value.(*object)
}

const (
	itemBody   = `{"properties": {"name": {"type": "string", "description": "its name"}}, "required": ["name"], "type": "object"}`
	itemParams = `[{"in": "path", "name": "itemID", "required": true, "schema": {"type": "string"}},
                 {"in": "query", "name": "force", "schema": {"type": "boolean"}}]`
)

func fingerprintOf(t *testing.T, doc *object) string {
	t.Helper()
	ops, err := operationFingerprints(doc)
	assert.NoError(t, err)
	assert.Equal(t, []string{"PUT /items/{itemID}"}, keysOf(ops), "only the write operations")
	return ops["PUT /items/{itemID}"]
}

func keysOf(m map[string]string) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func TestAWriteRequestsFingerprintIsWhatItCarries(t *testing.T) {
	base := fingerprintOf(t, specWith(t, itemBody, itemParams))

	assert.Equal(t, base, fingerprintOf(t, specWith(t,
		`{"properties": {"name": {"type": "string", "description": "another word"}}, "required": ["name"], "type": "object"}`,
		itemParams)), "a description is not what a request carries")
	assert.Equal(t, base, fingerprintOf(t, specWith(t, itemBody,
		`[{"in": "query", "name": "force", "schema": {"type": "boolean"}},
		  {"in": "path", "name": "itemID", "required": true, "schema": {"type": "string"}}]`)),
		"nor the order the parameters are documented in")

	assert.NotEqual(t, base, fingerprintOf(t, specWith(t,
		`{"properties": {"name": {"type": "string"}, "note": {"type": "string"}}, "required": ["name"], "type": "object"}`,
		itemParams)), "a field added to the body, through its reference")
	assert.NotEqual(t, base, fingerprintOf(t, specWith(t, itemBody,
		`[{"in": "path", "name": "itemID", "required": true, "schema": {"type": "string"}}]`)), "a parameter gone")
}

func TestTheAPILevelRisesWhenAnExistingWriteChanges(t *testing.T) {
	first, changed := nextAPILevel(apiLevel{}, map[string]string{"PUT /a": "1"})
	assert.Equal(t, 1, first.Level, "a first run starts at 1")
	assert.Empty(t, changed)

	same, _ := nextAPILevel(first, map[string]string{"PUT /a": "1"})
	assert.Equal(t, 1, same.Level)

	added, _ := nextAPILevel(first, map[string]string{"PUT /a": "1", "POST /b": "2"})
	assert.Equal(t, 1, added.Level, "an operation an older CLI never calls asks nothing of it")

	removed, _ := nextAPILevel(added, map[string]string{"POST /b": "2"})
	assert.Equal(t, 1, removed.Level)

	raised, changed := nextAPILevel(added, map[string]string{"PUT /a": "9", "POST /b": "2"})
	assert.Equal(t, 2, raised.Level)
	assert.Equal(t, []string{"PUT /a"}, changed)
}
