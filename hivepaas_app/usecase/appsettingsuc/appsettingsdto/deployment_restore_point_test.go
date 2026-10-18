package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The database a deployment marks a point on is saved as asked, and left out
// it is none: the request is the whole of the settings.
func TestTheRestorePointDatabaseIsSavedAsAsked(t *testing.T) {
	toEntity := func(database *basedto.ObjectIDReq) entity.ObjectID {
		req := &DeploymentSettingsReq{ActiveMethod: base.DeploymentMethodImage,
			ImageSource: &DeploymentImageSourceReq{Enabled: true, Image: "nginx:1"}, RestorePointBeforeDeploy: database}
		settings, err := req.ToEntity()
		assert.NoError(t, err)
		return settings.RestorePointBeforeDeploy
	}

	assert.Equal(t, entity.ObjectID{ID: "db-1"}, toEntity(&basedto.ObjectIDReq{ID: "db-1"}))
	assert.Equal(t, entity.ObjectID{}, toEntity(nil))
}

func TestTheSettingsSayWhichDatabaseIsMarked(t *testing.T) {
	transform := func(settings *entity.AppDeploymentSettings) *DeploymentSettingsResp {
		setting := &entity.Setting{ID: "setting-1", Type: base.SettingTypeAppDeployment}
		assert.NoError(t, setting.SetData(settings))
		resp, err := TransformDeploymentSettings(&AppDeploymentSettingsTransformInput{
			DeploymentSettings: setting, RefObjects: entity.NewRefObjects()})
		assert.NoError(t, err)
		return resp
	}

	marked := transform(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage,
		RestorePointBeforeDeploy: entity.ObjectID{ID: "db-1"}})
	assert.Equal(t, &basedto.ObjectIDResp{ID: "db-1"}, marked.RestorePointBeforeDeploy)
	assert.Nil(t, transform(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage}).
		RestorePointBeforeDeploy)
}
