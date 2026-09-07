package approutingserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
)

// Every case here returns before the service touches a dependency, which is the
// point: a revert that must not happen must not reach the database or docker at
// all. A zero-value service is enough, and it is also the assertion - anything
// that starts touching deps in these paths panics instead of passing quietly.
func TestRevertSettingsRefusals(t *testing.T) {
	ctx := context.Background()
	appWith := func(setting *entity.Setting) *entity.App {
		return &entity.App{Settings: []*entity.Setting{setting}}
	}
	goodSnapshot := entity.SettingSnapshot{Data: `{"exposePublicly":true}`, Version: 1}

	t.Run("a setting that has moved on is left alone", func(t *testing.T) {
		// The change under trial was confirmed and replaced by another one. Acting
		// now would overwrite settings this task has never seen.
		app := appWith(&entity.Setting{ID: "set-1", UpdateVer: 9})
		resp, err := (&service{}).RevertSettings(ctx, database.Tx{}, &approutingservice.RevertSettingsReq{
			App:          app,
			SettingID:    "set-1",
			ProbationVer: 7,
			Snapshot:     goodSnapshot,
		})
		assert.NoError(t, err)
		assert.False(t, resp.Reverted)
		assert.Equal(t, "settings changed since", resp.Reason)
		assert.Equal(t, 9, app.Settings[0].UpdateVer, "the setting must not be written")
	})

	t.Run("a setting that no longer exists", func(t *testing.T) {
		resp, err := (&service{}).RevertSettings(ctx, database.Tx{}, &approutingservice.RevertSettingsReq{
			App:          appWith(&entity.Setting{ID: "other", UpdateVer: 7}),
			SettingID:    "set-1",
			ProbationVer: 7,
			Snapshot:     goodSnapshot,
		})
		assert.NoError(t, err)
		assert.False(t, resp.Reverted)
		assert.Equal(t, "setting no longer exists", resp.Reason)
	})

	t.Run("an empty snapshot is an error, never a write", func(t *testing.T) {
		// Restoring "" would blank the settings and take the app down for good,
		// which is worse than the change being reverted late or by hand.
		app := appWith(&entity.Setting{ID: "set-1", UpdateVer: 7, Data: `{"exposePublicly":true}`})
		_, err := (&service{}).RevertSettings(ctx, database.Tx{}, &approutingservice.RevertSettingsReq{
			App:          app,
			SettingID:    "set-1",
			ProbationVer: 7,
			Snapshot:     entity.SettingSnapshot{},
		})
		assert.Error(t, err)
		assert.Equal(t, `{"exposePublicly":true}`, app.Settings[0].Data)
	})
}

func TestFindSettingByID(t *testing.T) {
	app := &entity.App{Settings: []*entity.Setting{{ID: "a"}, {ID: "b"}}}
	assert.Equal(t, "b", findSettingByID(app, "b").ID)
	assert.Nil(t, findSettingByID(app, "c"))
	assert.Nil(t, findSettingByID(&entity.App{}, "a"))
}
