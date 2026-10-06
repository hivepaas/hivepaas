package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// The system apps are synced unless switched off: a setting saved before the
// switch existed syncs too.
func TestSystemAppsSyncIsOnUnlessSwitchedOff(t *testing.T) {
	assert.True(t, (&SystemCleanup{}).SystemAppsSyncEnabled())
	assert.True(t, (&SystemCleanup{SystemAppsSync: &SystemAppsSync{Enabled: true}}).SystemAppsSyncEnabled())
	assert.False(t, (&SystemCleanup{SystemAppsSync: &SystemAppsSync{}}).SystemAppsSyncEnabled())
}

// Version 1 gains the switch, on; one already set is kept.
func TestSystemCleanupMigrationSwitchesTheSyncOn(t *testing.T) {
	setting := &Setting{Type: base.SettingTypeSystemCleanup, Version: 1, Data: `{"fileCleanup":{"enabled":true}}`}
	changed, err := setting.Migrate()
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, CurrentSystemCleanupVersion, setting.Version)
	assert.Contains(t, setting.Data, `"systemAppsSync":{"enabled":true}`)

	off := &Setting{Type: base.SettingTypeSystemCleanup, Version: 1, Data: `{"systemAppsSync":{"enabled":false}}`}
	_, err = off.Migrate()
	assert.NoError(t, err)
	assert.Contains(t, off.Data, `"systemAppsSync":{"enabled":false}`)
}
