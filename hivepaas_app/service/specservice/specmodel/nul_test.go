package specmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNULSaysWhereByTheDocumentsNames(t *testing.T) {
	_, found := FindNUL(&EnvDoc{Name: "dev", Apps: map[string]*AppDoc{"web": {Name: "Web"}}})
	assert.False(t, found)
	_, found = FindNUL(nil)
	assert.False(t, found)
	_, found = FindNUL("a\x00")
	assert.True(t, found, "at the root too")

	for want, doc := range map[string]any{
		"name":          &EnvDoc{Name: "d\x00ev"},
		"apps.web.note": &EnvDoc{Apps: map[string]*AppDoc{"web": {Note: "a\x00"}}},
		"apps.web.settings.envVars.data[1].v": &EnvDoc{Apps: map[string]*AppDoc{"web": {
			Settings: map[string]any{"envVars": map[string]any{"data": []any{
				map[string]any{"v": "ok"}, map[string]any{"v": "x\x00"},
			}}},
		}}},
		"settings.a\x00": &ProjectDoc{Settings: map[string]any{"a\x00": 1}},
	} {
		at, found := FindNUL(doc)
		assert.True(t, found, want)
		assert.Equal(t, want, at)
	}
}
