package appcloneserviceimpl

import (
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appcloneservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func onProd(t *testing.T) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: config.EnvProd})
	t.Cleanup(func() { config.SetCurrent(previous) })
}

func withoutDeploymentSettings() *appCloneData {
	return &appCloneData{AppCloneReq: &appcloneservice.AppCloneReq{CloneSettings: &entity.AppCloneSettings{}}}
}

// A clone that does not take the deployment settings starts as a new app does:
// on the placeholder, with no command, in its settings and in its service.
func TestACloneWithoutDeploymentSettingsStartsOnThePlaceholder(t *testing.T) {
	onProd(t)
	placeholder := systemappservice.CurrentRelease().PlaceholderImage
	setting := &entity.Setting{Type: base.SettingTypeAppDeployment}
	assert.NoError(t, setting.SetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage, ImageSource: &entity.DeploymentImageSource{Image: "nginx:1"},
		Command: "nginx -g 'daemon off;'", WorkingDir: "/srv"}))

	cloned, err := (&service{}).onCloneDeploymentSettingDefault(setting, withoutDeploymentSettings())

	assert.NoError(t, err)
	settings := cloned.MustAsAppDeploymentSettings()
	assert.Equal(t, placeholder, settings.ImageSource.Image)
	assert.Empty(t, settings.Command)
	assert.Empty(t, settings.WorkingDir)

	svc := &swarm.Service{Spec: swarm.ServiceSpec{TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{
		Image: "nginx:1", Command: []string{"nginx"}, Args: []string{"-g", "daemon off;"}, Dir: "/srv"}}}}
	assert.NoError(t, (&service{}).onCloneServiceDefault(svc, nil, withoutDeploymentSettings()))
	container := svc.Spec.TaskTemplate.ContainerSpec
	assert.Equal(t, placeholder, container.Image)
	assert.Empty(t, container.Command)
	assert.Empty(t, container.Args)
	assert.Empty(t, container.Dir)
}

// A clone marks no restore point before its deploys until it is asked to, even
// when it takes the deployment settings: the database is its source's, and may
// be of an environment the clone is not in.
func TestACloneMarksNoRestorePointUntilAsked(t *testing.T) {
	onProd(t)
	for name, data := range map[string]*appCloneData{
		"without the deployment settings": withoutDeploymentSettings(),
		"with them": {AppCloneReq: &appcloneservice.AppCloneReq{
			CloneSettings: &entity.AppCloneSettings{CloneDeploymentSettings: true}}},
	} {
		setting := &entity.Setting{Type: base.SettingTypeAppDeployment}
		assert.NoError(t, setting.SetData(&entity.AppDeploymentSettings{
			ActiveMethod: base.DeploymentMethodImage, ImageSource: &entity.DeploymentImageSource{Image: "nginx:1"},
			RestorePointBeforeDeploy: entity.ObjectID{ID: "db-1"}}))

		cloned, err := (&service{}).onCloneDeploymentSettingDefault(setting, data)

		assert.NoError(t, err, name)
		assert.Empty(t, cloned.MustAsAppDeploymentSettings().RestorePointBeforeDeploy.ID, name)
	}
}
