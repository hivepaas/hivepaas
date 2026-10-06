package registryserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// registryRow is the settings with the registry's row, or none.
type registryRow struct {
	memSettingRepo
	row *entity.Setting
}

func (r *registryRow) GetSingle(context.Context, database.IDB, *entity.ObjectScope, base.SettingType, bool,
	...bunex.SelectQueryOption) (*entity.Setting, error) {
	if r.row == nil {
		return nil, hperrors.Wrap(hperrors.ErrNotFound)
	}
	return r.row, nil
}

// removingApps is fakeSystemApps whose removal is seen: the app is gone after
// it, and its storage is said to stay or go.
type removingApps struct {
	fakeSystemApps
	removedStorage bool
}

func (f *removingApps) Remove(_ context.Context, _ database.IDB, _ *entity.App, removeStorage bool) error {
	f.removed, f.removedStorage, f.app = true, removeStorage, nil
	return nil
}

// Switched off - by a save that did not confirm, or a row edited - or not
// configured at all, the registry's app is removed with its images kept, and
// the credential it pushed with is not offered for deletion.
func TestSyncRemovesARegistryTheSettingsNoLongerWant(t *testing.T) {
	for _, tc := range []struct {
		name string
		row  func(t *testing.T) *entity.Setting
	}{
		{"switched off", func(t *testing.T) *entity.Setting {
			return registrySetting(t, &entity.RegistrySettings{AppID: "app-1", RegistryAuthID: "cred-1"})
		}},
		{"not configured", func(*testing.T) *entity.Setting { return nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			apps := &removingApps{fakeSystemApps: fakeSystemApps{app: &entity.App{ID: "app-1", Name: "Registry"}}}
			repo := &registryRow{memSettingRepo: memSettingRepo{settings: map[string]*entity.Setting{}}, row: tc.row(t)}
			s := &service{settingRepo: repo, systemAppService: apps, logger: quietLogger{}}

			resp, err := s.Sync(context.Background(), nil)

			assert.NoError(t, err)
			assert.True(t, apps.removed)
			assert.False(t, apps.removedStorage, "the images stay")
			if assert.NotNil(t, resp.App) {
				assert.Equal(t, entity.SystemAppSyncRemoved, resp.App.Action)
				assert.Equal(t, "app-1", resp.App.AppID)
			}
			assert.Empty(t, resp.Tasks)
		})
	}
}

// Nothing configured and nothing running: nothing to say.
func TestSyncWithNoRegistrySaysNothing(t *testing.T) {
	repo := &registryRow{memSettingRepo: memSettingRepo{settings: map[string]*entity.Setting{}}}
	s := &service{settingRepo: repo, systemAppService: &removingApps{}, logger: quietLogger{}}

	resp, err := s.Sync(context.Background(), nil)

	assert.NoError(t, err)
	assert.Nil(t, resp.App)
}
