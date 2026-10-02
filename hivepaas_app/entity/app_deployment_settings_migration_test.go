package entity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// A row written before this change carries a name, tags and a no-cache flag that
// nothing reads any more. The migration drops them, so that a spec bundle
// exported afterwards does not carry them either.
func TestDeploymentSettingsMigrationDropsNamingFields(t *testing.T) {
	stored := &entity.Setting{
		ID:      "s1",
		Type:    base.SettingTypeAppDeployment,
		Version: 1,
		Data: `{
			"activeMethod": "repo",
			"repoSource": {
				"repoType": "git", "repoURL": "https://example.com/x.git", "repoRef": "main",
				"imageName": "custom-name", "imageTags": ["api:latest"]
			},
			"noCache": true
		}`,
	}

	data, err := stored.AsAppDeploymentSettings()
	assert.NoError(t, err)

	changed, err := data.Migrate(stored)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, entity.CurrentAppDeploymentSettingsVersion, stored.Version)

	assert.NotContains(t, stored.Data, "imageName")
	assert.NotContains(t, stored.Data, "custom-name")
	assert.NotContains(t, stored.Data, "api:latest")
	assert.NotContains(t, stored.Data, "noCache")
	assert.Contains(t, stored.Data, "repoRef")
}

// An app built from a repository deployed on every push before it could be told
// not to: the migration keeps it doing so. A function did not, and still does
// not until it is told to.
func TestDeploymentSettingsMigrationKeepsAppsDeployingOnPush(t *testing.T) {
	app := &entity.Setting{
		ID: "s1", Type: base.SettingTypeAppDeployment, Version: 2,
		Data: `{"activeMethod": "repo", "repoSource": {"repoType": "git", "repoRef": "main"}}`,
	}
	data, err := app.AsAppDeploymentSettings()
	assert.NoError(t, err)

	changed, err := data.Migrate(app)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.True(t, app.MustAsAppDeploymentSettings().RepoSource.AutoDeploy)

	function := &entity.Setting{
		ID: "s2", Type: base.SettingTypeAppDeployment, Version: 2,
		Data: `{"activeMethod": "function", "functionSource": {"code": {"repo": {"repoRef": "main"}}}}`,
	}
	data, err = function.AsAppDeploymentSettings()
	assert.NoError(t, err)

	_, err = data.Migrate(function)
	assert.NoError(t, err)
	assert.False(t, function.MustAsAppDeploymentSettings().FunctionSource.Code.Repo.AutoDeploy)
}
