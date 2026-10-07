package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// settings.data is JSONB, and Postgres refuses the \u0000 JSON writes for a NUL:
// the statement fails, and the transaction it is in with it - a snapshot sync, a
// volume sync, an import. The NUL is dropped instead, as a task's args and
// output drop it, and what is read afterwards is what was stored.
func TestSettingDataDropsNULCharacters(t *testing.T) {
	s := &Setting{Type: base.SettingTypeEnvVar}

	assert.NoError(t, s.SetData(&EnvVars{Data: []*EnvVar{{Key: "A\x00", Value: "x\x00y"}}}))

	assert.NotContains(t, s.Data, `\u0000`)
	assert.Equal(t, int64(len(s.Data)), s.Size)
	vars := s.MustAsEnvVars()
	assert.Equal(t, "A", vars.Data[0].Key)
	assert.Equal(t, "xy", vars.Data[0].Value)
}

// A backslash before "u0000" is text, not a NUL: kept as it is.
func TestSettingDataKeepsTheTextOfAnEscape(t *testing.T) {
	s := &Setting{Type: base.SettingTypeEnvVar}
	data := &EnvVars{Data: []*EnvVar{{Key: "A", Value: `\u0000`}}}

	assert.NoError(t, s.SetData(data))

	assert.Same(t, data, s.MustAsEnvVars(), "nothing dropped: the data set is the data read")
	assert.Equal(t, `\u0000`, s.MustAsEnvVars().Data[0].Value)
}
