package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func decoded(t *testing.T, s string) any {
	t.Helper()
	var v any
	if !assert.NoError(t, json.Unmarshal([]byte(s), &v)) {
		t.FailNow()
	}
	return v
}

// The RFC 7396 examples that matter here: fields merge, null removes, and a
// list is replaced whole.
func TestMergePatch(t *testing.T) {
	for _, c := range []struct{ doc, patch, want string }{
		{`{"a":"b"}`, `{"a":"c"}`, `{"a":"c"}`},
		{`{"a":"b"}`, `{"b":"c"}`, `{"a":"b","b":"c"}`},
		{`{"a":"b"}`, `{"a":null}`, `{}`},
		{`{"a":{"b":"c","d":"e"}}`, `{"a":{"b":null,"f":"g"}}`, `{"a":{"d":"e","f":"g"}}`},
		{`{"a":[1,2]}`, `{"a":[3]}`, `{"a":[3]}`},
		{`{"a":"b"}`, `{"a":{"c":"d"}}`, `{"a":{"c":"d"}}`},
	} {
		assert.Equal(t, decoded(t, c.want), mergePatch(decoded(t, c.doc), decoded(t, c.patch)), c.patch)
	}
}

func TestDiffFieldsNamesEachValue(t *testing.T) {
	before := decoded(t, `{"limits":{"memory":"512MB","cpus":1},"envs":[{"key":"A","value":"1"}],"updateVer":3}`)
	after := decoded(t, `{"limits":{"memory":"1GB","cpus":1},"envs":[{"key":"A","value":"2"}],"updateVer":3}`)
	assert.Equal(t, []fieldChange{
		{Path: "envs", Before: decoded(t, `[{"key":"A","value":"1"}]`), After: decoded(t, `[{"key":"A","value":"2"}]`)},
		{Path: "limits.memory", Before: "512MB", After: "1GB"},
	}, diffFields(before, after))
	assert.Empty(t, diffFields(before, before))
}

func TestDroppedFieldsAreWhatTheEndpointDoesNotTake(t *testing.T) {
	assert.Equal(t, []string{"inheritedRuntimeEnvVars"},
		droppedFields(map[string]any{"runtimeEnvVars": []any{}, "inheritedRuntimeEnvVars": []any{}, "gone": nil},
			map[string]any{"runtimeEnvVars": []any{}}))
}
