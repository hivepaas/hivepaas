package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A stored false is written, not left out, so what an administrator turned off
// reads back off.
func TestMCPSettingsSurvivePersistence(t *testing.T) {
	for _, want := range []*entity.MCPSettings{{Enabled: true}, {Enabled: false, AllowWrite: true}, {}} {
		s := &entity.Setting{ID: "s1", Type: base.SettingTypeMCP}
		if !assert.NoError(t, s.SetData(want)) {
			t.FailNow()
		}
		stored := &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
		got, err := stored.AsMCPSettings()
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

func TestMCPSettingsAreOffUntilSet(t *testing.T) {
	got, err := (&entity.Setting{Type: base.SettingTypeMCP, Data: "{}"}).AsMCPSettings()
	assert.NoError(t, err)
	assert.False(t, got.Enabled)
}
