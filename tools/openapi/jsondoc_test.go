package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Reading the committed spec and writing it back changes nothing: the fixes
// are the only difference a run makes.
func TestEncodeDocumentWritesTheSpecAsTheConverterDoes(t *testing.T) {
	data, err := os.ReadFile("../../docs/openapi/swagger.json")
	assert.NoError(t, err)

	doc, err := decodeDocument(data)

	assert.NoError(t, err)
	assert.Equal(t, string(data), string(encodeDocument(doc)))
}

func TestEncodeDocument(t *testing.T) {
	doc, err := decodeDocument([]byte(`{"a":[],"b":{},"c":[{"d":1.50,"e":"x\ny\u0001é"}],"f":[1,2],"g":null,"h":true}`))
	assert.NoError(t, err)

	assert.Equal(t, `{
  "a" : [ ],
  "b" : { },
  "c" : [ {
    "d" : 1.50,
    "e" : "x\ny\u0001é"
  } ],
  "f" : [ 1, 2 ],
  "g" : null,
  "h" : true
}`, string(encodeDocument(doc)))
}
