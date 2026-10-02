package githubappdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func pathsOf(errs hperrors.ValidationErrors) []string {
	var paths []string
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return paths
}

func createReq(base GithubAppBaseReq) *CreateGithubAppReq {
	req := NewCreateGithubAppReq()
	req.GithubAppBaseReq = &base
	return req
}

// Reading a repository through the app takes its id, its installation and its
// private key; nothing else.
func TestAGithubAppReadsReposWithItsIDInstallationAndKey(t *testing.T) {
	req := createReq(GithubAppBaseReq{Name: "acme", GhAppID: 1, GhInstallationID: 2, PrivateKey: "-----BEGIN KEY-----"})

	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, pathsOf(req.Validate()))
}

func TestAGithubAppWithoutItsIDInstallationOrKeyIsRefused(t *testing.T) {
	req := createReq(GithubAppBaseReq{Name: "acme"})

	assert.NoError(t, req.ModifyRequest())
	assert.ElementsMatch(t, []string{"appId", "installationId", "privateKey"}, pathsOf(req.Validate()))
}

// Signing in through the app is OAuth: it takes the client's id and secret.
func TestSignInThroughAGithubAppTakesItsClient(t *testing.T) {
	req := createReq(GithubAppBaseReq{Name: "acme", GhAppID: 1, GhInstallationID: 2, PrivateKey: "key",
		SSOEnabled: true})

	assert.NoError(t, req.ModifyRequest())
	assert.ElementsMatch(t, []string{"clientId", "clientSecret"}, pathsOf(req.Validate()))
}
