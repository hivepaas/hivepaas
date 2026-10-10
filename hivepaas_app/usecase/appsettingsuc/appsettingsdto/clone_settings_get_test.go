package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func commandPipeSetting(id, name string) *entity.Setting {
	return &entity.Setting{ID: id, Type: base.SettingTypeCommandPipe, Name: name, Status: base.SettingStatusActive}
}

// The clone's command pipes are listed once each, named: saved twice - as the
// form did, given each twice - they are still listed once.
func TestTheCloneSettingsListEachCommandPipeOnce(t *testing.T) {
	refObjects := entity.NewRefObjects()
	refObjects.RefSettings["pipe-1"] = commandPipeSetting("pipe-1", "dump-db")
	refObjects.RefSettings["pipe-2"] = commandPipeSetting("pipe-2", "copy-files")

	resp, err := TransformAppCloneSettings(&AppCloneSettingsTransformInput{
		App: &entity.App{ID: "app-1"},
		AppCloneSettings: &entity.AppCloneSettings{
			CommandPipes: entity.ObjectIDSlice{{ID: "pipe-1"}, {ID: "pipe-2"}, {ID: "pipe-1"}},
		},
		RefObjects: refObjects,
	})
	assert.NoError(t, err)

	names := make([]string, 0, len(resp.CommandPipes))
	for _, pipe := range resp.CommandPipes {
		names = append(names, pipe.Name)
	}
	assert.Equal(t, []string{"dump-db", "copy-files"}, names)
}

// A save keeps each pipe once, in the order given.
func TestTheCloneSettingsSaveEachCommandPipeOnce(t *testing.T) {
	req := &UpdateAppCloneSettingsReq{
		CommandPipes: basedto.ObjectIDSliceReq{{ID: "pipe-2"}, {ID: "pipe-1"}, {ID: "pipe-2"}},
	}

	assert.Equal(t, []string{"pipe-2", "pipe-1"}, req.ToEntity().CommandPipes.ToIDStringSlice())
}
