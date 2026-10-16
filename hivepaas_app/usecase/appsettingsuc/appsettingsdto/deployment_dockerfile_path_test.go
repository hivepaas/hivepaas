package appsettingsdto

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

// The Dockerfile's path and the scan path are places in the repository: one
// that leads out of it by ".." is refused when it is saved.
func TestDockerfilePathsStayInTheRepository(t *testing.T) {
	invalid := func(path, scanPath string) []string {
		req := imageDeploymentReq()
		req.RepoSource = &DeploymentRepoSourceReq{
			RepoType: base.RepoTypeGit, RepoURL: "https://example.com/r.git", RepoRef: "main",
			Dockerfile: DeploymentDockerfileReq{Source: base.DockerfileSourceAuto, Path: path, ScanPath: scanPath},
		}
		var found []string
		for _, inner := range req.Validate().Build(translation.LangEn).InnerErrors {
			if slices.Contains([]string{"repoSource.dockerfile.path", "repoSource.dockerfile.scanPath"}, inner.Path) {
				found = append(found, inner.Path)
			}
		}
		return found
	}

	for _, ok := range []string{"", "Dockerfile", "/Dockerfile", "docker/Dockerfile.prod", "apps/../Dockerfile"} {
		assert.Empty(t, invalid(ok, ok), ok)
	}
	for _, bad := range []string{"..", "../Dockerfile", "apps/../../Dockerfile", "//etc/cron.d/x"} {
		assert.Equal(t, []string{"repoSource.dockerfile.path", "repoSource.dockerfile.scanPath"},
			invalid(bad, bad), bad)
	}
}
