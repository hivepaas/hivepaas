package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func functionSettings() *AppDeploymentSettings {
	return &AppDeploymentSettings{
		FunctionSource: &DeploymentFunctionSource{
			Runtime: base.FunctionRuntimeNode24,
			Code: FunctionCode{Repo: &FunctionRepoCode{
				RepoType: base.RepoTypeGit, RepoID: "github.com/acme/fns", RepoURL: "https://github.com/acme/fns.git",
				RepoRef: "refs/heads/main", CommitHash: "abc123",
				RepoOptions: DeploymentRepoOptions{GitSubmodulesEnabled: true},
				Credentials: RepoCredentials{ID: "cred-1", Type: base.SettingTypeGithubApp},
			}},
			PushToRegistry: ObjectID{ID: "registry-1"},
		},
	}
}

// A function's registry and its repository's credentials are settings it
// refers to, as an app's are: loaded with it, and kept from being deleted.
func TestAFunctionsRegistryAndCredentialsAreItsReferences(t *testing.T) {
	settings := functionSettings()

	assert.ElementsMatch(t, []string{"registry-1"}, settings.GetRegistryAuthIDs())
	assert.ElementsMatch(t, []string{"cred-1"}, settings.GetGitCredentialIDs())
	assert.ElementsMatch(t, []string{"registry-1", "cred-1"}, settings.GetRefObjectIDs().RefSettingIDs)
}

// A function's repository is checked out as an app's is.
func TestAFunctionsRepoIsCheckedOutAsARepoSource(t *testing.T) {
	code := functionSettings().FunctionSource.Code.Repo

	source := code.RepoSource()

	assert.Equal(t, &DeploymentRepoSource{
		RepoType: base.RepoTypeGit, RepoID: "github.com/acme/fns", RepoURL: "https://github.com/acme/fns.git",
		RepoRef: "refs/heads/main", CommitHash: "abc123",
		RepoOptions: DeploymentRepoOptions{GitSubmodulesEnabled: true},
		Credentials: RepoCredentials{ID: "cred-1", Type: base.SettingTypeGithubApp},
	}, source)
}

// Inline code is kept in the setting, so the setting reads back the same.
func TestAFunctionsInlineCodeIsKeptInItsSetting(t *testing.T) {
	settings := &AppDeploymentSettings{
		FunctionSource: &DeploymentFunctionSource{
			Runtime: base.FunctionRuntimePython313,
			Code: FunctionCode{Inline: &FunctionInlineCode{Files: []*FunctionFile{
				{Path: "main.py", Content: "def handler(req, ctx):\n    return None\n"},
			}}},
		},
	}
	setting := &Setting{Type: base.SettingTypeAppDeployment}

	assert.NoError(t, setting.SetData(settings))
	read, err := setting.AsAppDeploymentSettings()

	assert.NoError(t, err)
	assert.Equal(t, settings.FunctionSource, read.FunctionSource)
}
