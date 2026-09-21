package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestTransformAppKindSettings_NilKindSetting(t *testing.T) {
	input := &AppKindSettingsTransformInput{
		App:         &entity.App{ID: "app-1"},
		KindSetting: nil,
	}

	resp, err := TransformAppKindSettings(input)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, base.AppCategoryWebapp, resp.Category)
	assert.Equal(t, uint(0), resp.Port)
	assert.Equal(t, "", resp.Engine)
	assert.Nil(t, resp.Database)
	assert.Nil(t, resp.Cache)
	assert.Nil(t, resp.Storage)
}

func TestTransformAppKindSettings_WithRoutingPortFallback(t *testing.T) {
	routingSetting := &entity.Setting{
		ID:   "set-routing",
		Type: base.SettingTypeAppRouting,
	}
	routingSetting.MustSetData(&entity.AppRoutingSettings{
		Port: 3000,
	})

	input := &AppKindSettingsTransformInput{
		App:            &entity.App{ID: "app-1"},
		KindSetting:    nil,
		RoutingSetting: routingSetting,
	}

	resp, err := TransformAppKindSettings(input)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, base.AppCategoryWebapp, resp.Category)
	assert.Equal(t, uint(3000), resp.Port)
}

func TestTransformAppKindSettings_WithKindSetting(t *testing.T) {
	kindSetting := &entity.Setting{
		ID:        "set-kind",
		Type:      base.SettingTypeAppKind,
		UpdateVer: 2,
		Version:   1,
	}
	kindSetting.MustSetData(&entity.AppKindSettings{
		Category: base.AppCategoryDatabase,
		Engine:   "postgres",
		Version:  "16",
	})

	input := &AppKindSettingsTransformInput{
		App:         &entity.App{ID: "app-1"},
		KindSetting: kindSetting,
	}

	resp, err := TransformAppKindSettings(input)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, base.AppCategoryDatabase, resp.Category)
	assert.Equal(t, "postgres", resp.Engine)
	assert.Equal(t, "16", resp.Version)
	assert.Equal(t, 2, resp.UpdateVer)
}
