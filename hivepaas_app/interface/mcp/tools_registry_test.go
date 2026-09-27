package mcp

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every tool is served, and says what each of its arguments is: a model reads
// nothing else about them.
func TestEveryToolBuildsAndDescribesItsArguments(t *testing.T) {
	w := newWriteWorld(t)
	listed, err := w.session(t, "key1").ListTools(context.Background(), nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	served := make([]string, 0, len(listed.Tools))
	for _, tool := range listed.Tools {
		served = append(served, tool.Name)
		assert.NotEmpty(t, tool.Description, tool.Name)
		assert.Nil(t, tool.OutputSchema, "%s: an answer is the API's own, with no schema", tool.Name)

		raw, err := json.Marshal(tool.InputSchema)
		if !assert.NoError(t, err) {
			continue
		}
		var schema struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
			Required []string `json:"required"`
		}
		assert.NoError(t, json.Unmarshal(raw, &schema), tool.Name)
		for name, prop := range schema.Properties {
			assert.NotEmpty(t, prop.Description, "%s: %s has no description", tool.Name, name)
		}
		for _, name := range schema.Required {
			assert.Contains(t, schema.Properties, name, "%s requires what it does not take", tool.Name)
		}
	}
	for _, tool := range Tools() {
		assert.Contains(t, served, tool.Name)
	}
}

// An endpoint tool describes only the query parameters its endpoint takes.
func TestEndpointToolsDescribeOnlyTheirParameters(t *testing.T) {
	names := map[string]bool{}
	for _, e := range getEndpoints() {
		assert.False(t, names[e.name], "%s is declared twice", e.name)
		names[e.name] = true
		params, err := e.queryParams()
		if !assert.NoError(t, err, e.name) {
			continue
		}
		for name := range e.params {
			assert.True(t, slices.ContainsFunc(params, func(p queryParam) bool { return p.name == name }),
				"%s describes %s, which its endpoint does not take", e.name, name)
		}
		for _, name := range e.omit {
			assert.NotContains(t, e.params, name, e.name)
		}
	}
}
