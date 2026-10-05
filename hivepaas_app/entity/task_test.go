package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A task's output is stored as JSONB, which refuses \u0000: an error text
// holding a NUL has it dropped, and what the output parses to is read from
// what was stored.
func TestATaskOutputDropsTheNULsJSONBRefuses(t *testing.T) {
	task := &Task{}
	assert.NoError(t, task.SetOutput(&TaskBackupRepoCleanupOutput{ReposFailed: 1,
		Repos: []*TaskBackupRepoCleanupRepoOutput{{RepoID: "r1", Error: "kopia: \x00bad\x00 header"}}}))
	assert.NotContains(t, task.Output, `\u0000`)
	assert.Contains(t, task.Output, `kopia: bad header`)
	got, err := task.OutputAsBackupRepoCleanup()
	assert.NoError(t, err)
	if assert.Len(t, got.Repos, 1) {
		assert.Equal(t, "kopia: bad header", got.Repos[0].Error)
	}
	assert.Equal(t, 1, got.ReposFailed)
}

// The text "\u0000" - a backslash, then u0000 - is no NUL, and is kept; so is
// any number, exactly; and a clean value is stored as it marshals.
func TestMarshalJSONBKeepsWhatIsNoNUL(t *testing.T) {
	b, clean, err := marshalJSONB(map[string]any{"path": `C:\u0000dir`, "ns": int64(9007199254740993)})
	assert.NoError(t, err)
	assert.True(t, clean)
	assert.Equal(t, `{"ns":9007199254740993,"path":"C:\\u0000dir"}`, string(b))

	b, clean, err = marshalJSONB(map[string]any{"ns": int64(9007199254740993), "msg": "a\x00b",
		"list": []any{"c\x00"}, "k\x00ey": 1})
	assert.NoError(t, err)
	assert.False(t, clean)
	assert.False(t, strings.Contains(string(b), `\u0000`))
	assert.Equal(t, `{"key":1,"list":["c"],"msg":"ab","ns":9007199254740993}`, string(b))
}
