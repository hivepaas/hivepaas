package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func repoAppSettings(autoDeploy bool) *AppDeploymentSettings {
	return &AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodRepo,
		RepoSource: &DeploymentRepoSource{
			RepoType: base.RepoTypeGit, RepoID: "github.com/acme/api", RepoURL: "https://github.com/acme/api.git",
			RepoRef: "refs/heads/main", CommitHash: "abc123", AutoDeploy: autoDeploy,
		},
	}
}

func repoFunctionSettings(autoDeploy bool) *AppDeploymentSettings {
	settings := functionSettings()
	settings.ActiveMethod = base.DeploymentMethodFunction
	settings.FunctionSource.Code.Repo.AutoDeploy = autoDeploy
	return settings
}

// A push deploys an app built from that repository and ref, and only when the
// app is set to deploy on push.
func TestAPushDeploysWhatIsBuiltFromItsRepoAndRef(t *testing.T) {
	assert.True(t, repoAppSettings(true).DeploysOnPush("github.com/acme/api", "refs/heads/main"))
	assert.True(t, repoFunctionSettings(true).DeploysOnPush("github.com/acme/fns", "refs/heads/main"))

	assert.False(t, repoAppSettings(false).DeploysOnPush("github.com/acme/api", "refs/heads/main"))
	assert.False(t, repoFunctionSettings(false).DeploysOnPush("github.com/acme/fns", "refs/heads/main"))

	assert.False(t, repoAppSettings(true).DeploysOnPush("github.com/acme/other", "refs/heads/main"))
	assert.False(t, repoAppSettings(true).DeploysOnPush("github.com/acme/api", "refs/heads/dev"))
	assert.False(t, repoFunctionSettings(true).DeploysOnPush("github.com/acme/fns", "refs/heads/dev"))
}

// What a push to the repository does not build is not deployed by it: an image,
// a function's inline code, a repo source kept while another method is active.
func TestAPushDoesNotDeployWhatIsNotBuiltFromARepo(t *testing.T) {
	image := repoAppSettings(true)
	image.ActiveMethod = base.DeploymentMethodImage
	assert.False(t, image.DeploysOnPush("github.com/acme/api", "refs/heads/main"))

	inline := repoFunctionSettings(true)
	inline.FunctionSource.Code = FunctionCode{Inline: &FunctionInlineCode{}}
	assert.False(t, inline.DeploysOnPush("github.com/acme/fns", "refs/heads/main"))

	assert.False(t, (&AppDeploymentSettings{ActiveMethod: base.DeploymentMethodFunction}).
		DeploysOnPush("github.com/acme/fns", "refs/heads/main"))
}

// The commit a deployment builds is set on the source the app is built from.
func TestTheRepoCommitIsSetOnTheActiveSource(t *testing.T) {
	app := repoAppSettings(true)
	assert.True(t, app.SetRepoCommitHash("def456"))
	assert.Equal(t, "def456", app.RepoSource.CommitHash)
	assert.False(t, app.SetRepoCommitHash("def456"))

	fn := repoFunctionSettings(true)
	assert.True(t, fn.SetRepoCommitHash(""))
	assert.Empty(t, fn.FunctionSource.Code.Repo.CommitHash)

	image := &AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage}
	assert.False(t, image.SetRepoCommitHash("def456"))
}

// A function built from a repository is linked to it, as an app is, so that a
// push to it finds the function.
func TestAFunctionIsLinkedToItsRepo(t *testing.T) {
	setting := &Setting{ID: "setting-1", ObjectID: "app-1"}

	links := repoFunctionSettings(false).GetResourceLinks(setting)

	var repoIDs []string
	for _, link := range links {
		if link.DstType == base.ResourceTypeRepo {
			repoIDs = append(repoIDs, link.DstID)
		}
	}
	assert.Equal(t, []string{"github.com/acme/fns"}, repoIDs)
}
