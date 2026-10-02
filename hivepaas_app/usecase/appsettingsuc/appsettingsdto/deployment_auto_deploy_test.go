package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A repo source deploys on push unless the request says otherwise.
func TestARepoSourceDeploysOnPushUnlessToldNot(t *testing.T) {
	repoSource := func(autoDeploy *bool) *entity.DeploymentRepoSource {
		req := &DeploymentRepoSourceReq{
			RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/api.git", RepoRef: "main",
			AutoDeploy: autoDeploy,
		}
		source, err := req.ToEntity()
		assert.NoError(t, err)
		return source
	}

	assert.True(t, repoSource(nil).AutoDeploy)
	assert.True(t, repoSource(new(true)).AutoDeploy)
	assert.False(t, repoSource(new(false)).AutoDeploy)
}

// So does a function's repository: a new function built from one deploys on
// push, as an app does.
func TestAFunctionsRepoDeploysOnPushUnlessToldNot(t *testing.T) {
	repoCode := func(autoDeploy *bool) *entity.FunctionRepoCode {
		req := &DeploymentFunctionSourceReq{
			Runtime: base.FunctionRuntimeNode24,
			Code: FunctionCodeReq{Repo: &FunctionRepoCodeReq{
				RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/fns.git", RepoRef: "main",
				AutoDeploy: autoDeploy,
			}},
		}
		req.Normalize()
		source, err := req.ToEntity()
		assert.NoError(t, err)
		return source.Code.Repo
	}

	assert.True(t, repoCode(nil).AutoDeploy)
	assert.False(t, repoCode(new(false)).AutoDeploy)
}

// The settings say whether the app deploys on push.
func TestTheSettingsSayWhetherAPushDeploys(t *testing.T) {
	transform := func(settings *entity.AppDeploymentSettings) *DeploymentSettingsResp {
		setting := &entity.Setting{ID: "setting-1", Type: base.SettingTypeAppDeployment}
		assert.NoError(t, setting.SetData(settings))
		resp, err := TransformDeploymentSettings(&AppDeploymentSettingsTransformInput{
			DeploymentSettings: setting, RefObjects: entity.NewRefObjects()})
		assert.NoError(t, err)
		return resp
	}

	app := transform(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodRepo,
		RepoSource:   &entity.DeploymentRepoSource{RepoType: base.RepoTypeGit, AutoDeploy: true},
	})
	assert.True(t, app.RepoSource.AutoDeploy)

	fn := transform(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodFunction,
		FunctionSource: &entity.DeploymentFunctionSource{Code: entity.FunctionCode{
			Repo: &entity.FunctionRepoCode{RepoType: base.RepoTypeGit, AutoDeploy: true}}},
	})
	assert.True(t, fn.FunctionSource.Code.Repo.AutoDeploy)
}
