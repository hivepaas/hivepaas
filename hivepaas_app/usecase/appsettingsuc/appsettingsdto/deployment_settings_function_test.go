package appsettingsdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func functionSettingsReq(source *DeploymentFunctionSourceReq) *UpdateAppDeploymentSettingsReq {
	return &UpdateAppDeploymentSettingsReq{
		ProjectID:    "01J0000000000000000000PRJ1",
		ProjectEnvID: "01J0000000000000000000ENV1",
		AppID:        "01J0000000000000000000APP1",
		DeploymentSettingsReq: &DeploymentSettingsReq{
			ActiveMethod:   base.DeploymentMethodFunction,
			FunctionSource: source,
		},
	}
}

// A function's deployment settings carry its source, normalized on the way in.
func TestAFunctionsDeploymentSettingsCarryItsSource(t *testing.T) {
	req := functionSettingsReq(inlineSource(" python313 ", file("./main.py", "def handler(req, ctx): pass")))

	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, req.Validate())
	settings, err := req.ToEntity()

	assert.NoError(t, err)
	assert.Equal(t, base.FunctionRuntimePython313, settings.FunctionSource.Runtime)
	assert.Equal(t, "main.py", settings.FunctionSource.Code.Inline.Files[0].Path)
	assert.Equal(t, "handler", settings.FunctionSource.Entrypoint.Handler)
}

// pathsOf are the fields validation errors are on.
func pathsOf(errs hperrors.ValidationErrors) []string {
	var paths []string
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return paths
}

func TestTheFunctionMethodWithoutASourceIsRefused(t *testing.T) {
	req := functionSettingsReq(nil)

	assert.NoError(t, req.ModifyRequest())

	assert.Equal(t, []string{"functionSource"}, pathsOf(req.Validate()))
}

// A function's source is sent back with the settings: its code, and the
// settings it refers to as settings.
func TestAFunctionsSourceIsSentWithItsSettings(t *testing.T) {
	source := &entity.DeploymentFunctionSource{
		Runtime: base.FunctionRuntimeNode24, Contract: base.FunctionContractV1,
		Entrypoint: entity.FunctionEntrypoint{File: "index.js", Handler: "default"},
		Code: entity.FunctionCode{Dir: "fns/hello", Repo: &entity.FunctionRepoCode{
			RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/fns.git", RepoRef: "refs/heads/main",
			Credentials: entity.RepoCredentials{ID: "cred-1"},
		}},
		SystemPackages: []string{"ffmpeg"},
		MaxConcurrency: 16,
		PushToRegistry: entity.ObjectID{ID: "registry-1"},
	}
	setting := &entity.Setting{ID: "setting-1", Type: base.SettingTypeAppDeployment}
	assert.NoError(t, setting.SetData(&entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodFunction, FunctionSource: source}))
	registry := &entity.Setting{ID: "registry-1", Type: base.SettingTypeRegistryAuth, Name: "ghcr"}
	refObjects := entity.NewRefObjects()
	refObjects.RefSettings["registry-1"] = registry

	resp, err := TransformDeploymentSettings(&AppDeploymentSettingsTransformInput{
		DeploymentSettings: setting, RefObjects: refObjects})

	assert.NoError(t, err)
	fn := resp.FunctionSource
	assert.Equal(t, base.FunctionRuntimeNode24, fn.Runtime)
	assert.Equal(t, "index.js", fn.Entrypoint.File)
	assert.Equal(t, "fns/hello", fn.Code.Dir)
	assert.Equal(t, "https://github.com/acme/fns.git", fn.Code.Repo.RepoURL)
	assert.Equal(t, []string{"ffmpeg"}, fn.SystemPackages)
	assert.Equal(t, "ghcr", fn.PushToRegistry.Name)
	// A credential that is gone is shown as missing rather than dropped.
	assert.Equal(t, "cred-1", fn.Code.Repo.Credentials.ID)
}

// A function's kind is a function's, at the port its runtime listens on: a
// request cannot move it.
func TestAFunctionsKindKeepsItsRuntimesPort(t *testing.T) {
	req := &UpdateAppKindSettingsReq{
		ProjectID:          "01J0000000000000000000PRJ1",
		ProjectEnvID:       "01J0000000000000000000ENV1",
		AppID:              "01J0000000000000000000APP1",
		AppKindSettingsReq: &AppKindSettingsReq{Category: base.AppCategoryFunction, Port: 3000},
	}

	assert.NoError(t, req.ModifyRequest())

	assert.Equal(t, uint(base.FunctionPort), req.Port)
	assert.Empty(t, req.Validate())
}
