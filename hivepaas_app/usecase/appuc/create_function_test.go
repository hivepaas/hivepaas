package appuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

// fakeSettingService loads the settings a function's source refers to, and
// keeps where it was asked to look.
type fakeSettingService struct {
	settingservice.Service
	loaded map[string]*entity.Setting
	scope  *entity.ObjectScope
	refIDs *entity.RefObjectIDs
}

func (f *fakeSettingService) LoadRefObjectsByIDs(
	_ context.Context, _ database.IDB, refObjects **entity.RefObjects,
	scope *entity.ObjectScope, _ bool, refIDs *entity.RefObjectIDs,
) error {
	f.scope, f.refIDs = scope, refIDs
	if *refObjects == nil {
		*refObjects = entity.NewRefObjects()
	}
	for id, setting := range f.loaded {
		(*refObjects).RefSettings[id] = setting
	}
	return nil
}

// fakeDeploymentService keeps the build source it was asked to check.
type fakeDeploymentService struct {
	appdeploymentservice.Service
	checked *appdeploymentservice.CheckBuildSourceReq
}

func (f *fakeDeploymentService) CheckBuildSource(
	_ context.Context, req *appdeploymentservice.CheckBuildSourceReq,
) error {
	f.checked = req
	return nil
}

// A function's source is checked in its environment before anything is
// created: the settings it refers to, and what its build will need.
func TestAFunctionsSourceIsCheckedInItsEnvironmentFirst(t *testing.T) {
	credentials := &entity.Setting{ID: "cred-1", Type: base.SettingTypeAccessToken}
	settings := &fakeSettingService{loaded: map[string]*entity.Setting{"cred-1": credentials}}
	deployments := &fakeDeploymentService{}
	uc := &UC{settingService: settings, appDeploymentService: deployments}
	source := &entity.DeploymentFunctionSource{
		Runtime: base.FunctionRuntimeGo127,
		Code: entity.FunctionCode{Dir: "fns/hello", Repo: &entity.FunctionRepoCode{
			RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/fns.git", RepoRef: "refs/heads/main",
			Credentials: entity.RepoCredentials{ID: "cred-1"},
		}},
		PushToRegistry: entity.ObjectID{ID: "registry-1"},
	}

	err := uc.checkFunctionSource(context.Background(), "project-1", "env-1", source)

	assert.NoError(t, err)
	assert.Equal(t, base.ObjectScopeProjectEnv, settings.scope.ScopeType)
	assert.Equal(t, "project-1", settings.scope.ProjectID)
	assert.Equal(t, "env-1", settings.scope.ProjectEnvID)
	assert.ElementsMatch(t, []string{"cred-1", "registry-1"}, settings.refIDs.RefSettingIDs)
	assert.Equal(t, "https://github.com/acme/fns.git", deployments.checked.RepoSource.RepoURL)
	assert.Equal(t, "registry-1", deployments.checked.PushToRegistry.ID)
	assert.Same(t, credentials, deployments.checked.RefObjects.RefSettings["cred-1"])
}

// A function is created with its kind, its source to deploy, and its routing
// at the port its runtime listens on.
func TestAFunctionIsCreatedWithItsKindSourceAndPort(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	source := &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24}
	timeNow := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	settings := functionSettings(app, source, timeNow)

	byType := map[base.SettingType]*entity.Setting{}
	for _, setting := range settings {
		assert.Equal(t, "app-1", setting.ObjectID)
		assert.Equal(t, base.ObjectScopeApp, setting.Scope)
		assert.Equal(t, base.SettingStatusActive, setting.Status)
		assert.Equal(t, timeNow, setting.CreatedAt)
		assert.NotEmpty(t, setting.ID)
		byType[setting.Type] = setting
	}
	assert.Len(t, byType, 3)
	assert.True(t, entity.IsFunctionKind(byType[base.SettingTypeAppKind]))
	deployment := byType[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings()
	assert.Equal(t, base.DeploymentMethodFunction, deployment.ActiveMethod)
	assert.Equal(t, base.FunctionRuntimeNode24, deployment.FunctionSource.Runtime)
	assert.Equal(t, base.FunctionPort, byType[base.SettingTypeAppRouting].MustAsAppRoutingSettings().Port)
	// What a clone copies: its deployment and routing settings, as an app's.
	assert.True(t, byType[base.SettingTypeAppDeployment].Inheritable)
	assert.True(t, byType[base.SettingTypeAppRouting].Inheritable)
}
