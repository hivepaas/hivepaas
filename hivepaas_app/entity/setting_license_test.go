package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func TestLicenseSettingsSurvivePersistence(t *testing.T) {
	movedAt := time.Date(2026, 11, 2, 3, 4, 5, 0, time.UTC)
	refresh := entity.LicenseRefresh{
		Schedule: entity.SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day),
			InitialTime: time.Date(2026, 11, 1, 3, 30, 0, 0, time.UTC)},
		LastCheckedAt: movedAt.Add(time.Hour),
		LastResult:    base.LicenseRefreshFailed,
		LastError:     "The license portal could not be reached",
	}
	for _, want := range []*entity.LicenseSettings{{Key: "eyJrZXkiOjF9"}, {Key: "eyJrZXkiOjF9", MovedAt: movedAt}, {},
		{Key: "eyJrZXkiOjF9", Refresh: refresh}, {Refresh: entity.LicenseRefresh{Schedule: refresh.Schedule}}} {
		s := &entity.Setting{ID: "s1", Type: base.SettingTypeLicense}
		if !assert.NoError(t, s.SetData(want)) {
			t.FailNow()
		}
		stored := &entity.Setting{ID: s.ID, Type: s.Type, Data: s.Data}
		got, err := stored.AsLicenseSettings()
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}
}

// A key that was never moved writes no movedAt: a zero time would read back as
// a move in year one. A refresh that never ran writes nothing either.
func TestLicenseSettingsLeaveOutAMoveThatDidNotHappen(t *testing.T) {
	s := &entity.Setting{Type: base.SettingTypeLicense}
	if !assert.NoError(t, s.SetData(&entity.LicenseSettings{Key: "k"})) {
		t.FailNow()
	}
	assert.JSONEq(t, `{"key":"k"}`, s.Data)

	checkedAt := time.Date(2026, 11, 2, 3, 30, 0, 0, time.UTC)
	if !assert.NoError(t, s.SetData(&entity.LicenseSettings{Key: "k", Refresh: entity.LicenseRefresh{
		LastCheckedAt: checkedAt, LastResult: base.LicenseRefreshCurrent}})) {
		t.FailNow()
	}
	assert.JSONEq(t, `{"key":"k","refresh":{"lastCheckedAt":"2026-11-02T03:30:00Z","lastResult":"current"}}`,
		s.Data)
}

func TestLicenseSettingsHoldNoKeyUntilSet(t *testing.T) {
	got, err := (&entity.Setting{Type: base.SettingTypeLicense, Data: "{}"}).AsLicenseSettings()
	assert.NoError(t, err)
	assert.Empty(t, got.Key)
	assert.True(t, got.MovedAt.IsZero())
}

// A spec of one installation is applied to others, and a key names the one it
// is for; it also carries the customer's email.
func TestLicenseSettingsAreNeverExported(t *testing.T) {
	d := entity.SpecExportDecision(&entity.Setting{Type: base.SettingTypeLicense, Scope: base.ObjectScopeGlobal})
	assert.False(t, d.Export)
	assert.NotEmpty(t, d.Reason)
}

func TestLicenseSettingsMigration(t *testing.T) {
	current := &entity.Setting{Type: base.SettingTypeLicense, Version: entity.CurrentLicenseSettingsVersion,
		Data: `{"key":"k"}`}
	changed, err := current.Migrate()
	assert.NoError(t, err)
	assert.False(t, changed)

	unversioned := &entity.Setting{Type: base.SettingTypeLicense, Data: `{"key":"k"}`}
	changed, err = unversioned.Migrate()
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, entity.CurrentLicenseSettingsVersion, unversioned.Version)
	assert.Equal(t, "k", unversioned.MustAsLicenseSettings().Key)

	newer := &entity.Setting{Type: base.SettingTypeLicense, Version: entity.CurrentLicenseSettingsVersion + 1}
	_, err = newer.Migrate()
	assert.ErrorIs(t, err, hperrors.ErrDataVerNewerThanSystemVer)
}
