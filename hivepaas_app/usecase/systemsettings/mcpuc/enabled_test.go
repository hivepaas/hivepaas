package mcpuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type fakeSettingRepo struct {
	repository.SettingRepo
	setting *entity.Setting
	reads   int
}

func (f *fakeSettingRepo) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope,
	_ base.SettingType, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	f.reads++
	if f.setting == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return f.setting, nil
}

func mcpSetting(t *testing.T, enabled bool) *entity.Setting {
	t.Helper()
	s := &entity.Setting{Type: base.SettingTypeMCP, Status: base.SettingStatusActive,
		Version: entity.CurrentMCPSettingsVersion}
	if !assert.NoError(t, s.SetData(&entity.MCPSettings{Enabled: enabled})) {
		t.FailNow()
	}
	return s
}

func withClock(t *testing.T) *time.Time {
	t.Helper()
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	previous := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() { timeNow = previous })
	return &now
}

func TestTheServerIsOffUntilTurnedOn(t *testing.T) {
	withClock(t)
	uc := New(&settings.BaseUC{SettingRepo: &fakeSettingRepo{}})
	enabled, err := uc.IsEnabled(context.Background())
	assert.NoError(t, err)
	assert.False(t, enabled, "no setting saved")

	uc = New(&settings.BaseUC{SettingRepo: &fakeSettingRepo{setting: mcpSetting(t, false)}})
	enabled, _ = uc.IsEnabled(context.Background())
	assert.False(t, enabled)

	disabled := mcpSetting(t, true)
	disabled.Status = base.SettingStatusDisabled
	uc = New(&settings.BaseUC{SettingRepo: &fakeSettingRepo{setting: disabled}})
	enabled, _ = uc.IsEnabled(context.Background())
	assert.False(t, enabled, "a disabled setting serves nothing, whatever it says")
}

// Every MCP request asks; the database is asked once per enabledCacheTTL.
func TestTheSwitchIsReadOncePerInterval(t *testing.T) {
	now := withClock(t)
	repo := &fakeSettingRepo{setting: mcpSetting(t, true)}
	uc := New(&settings.BaseUC{SettingRepo: repo})

	for range 5 {
		enabled, err := uc.IsEnabled(context.Background())
		assert.NoError(t, err)
		assert.True(t, enabled)
	}
	assert.Equal(t, 1, repo.reads)

	repo.setting = mcpSetting(t, false)
	*now = now.Add(enabledCacheTTL)
	enabled, _ := uc.IsEnabled(context.Background())
	assert.False(t, enabled, "turned off elsewhere, seen within the interval")
	assert.Equal(t, 2, repo.reads)
}

func TestCurrentCarriesWhetherToolsMayWrite(t *testing.T) {
	withClock(t)
	s := &entity.Setting{Type: base.SettingTypeMCP, Status: base.SettingStatusActive,
		Version: entity.CurrentMCPSettingsVersion}
	if !assert.NoError(t, s.SetData(&entity.MCPSettings{Enabled: true, AllowWrite: true})) {
		t.FailNow()
	}
	uc := New(&settings.BaseUC{SettingRepo: &fakeSettingRepo{setting: s}})
	current, err := uc.Current(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, entity.MCPSettings{Enabled: true, AllowWrite: true}, current)
}
