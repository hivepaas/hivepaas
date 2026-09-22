package appactiondto_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
)

// reqWithTags is a request that would pass validation but for its tags, so that
// anything the tests see comes from the tags and not from a missing id.
func reqWithTags(tags ...string) *appactiondto.DeployAppReq {
	req := appactiondto.NewDeployAppReq()
	req.ProjectID = "01M2Z3M3B4BV4G2J6221ZH3Q29"
	req.ProjectEnvID = "01M2Z3M3B4BV4G2J6221ZH3Q29"
	req.AppID = "01M2Z3M3B4BV4G2J6221ZH3Q29"
	req.ActiveMethod = base.DeploymentMethodImage
	req.ImageTags = tags
	return req
}

// The helper has to be a request that passes on its own, or every case below
// would pass for the wrong reason.
func TestDeployWithoutTagsIsValid(t *testing.T) {
	assert.Empty(t, reqWithTags().Validate())
}

func TestDeployAcceptsPlainTags(t *testing.T) {
	assert.Empty(t, reqWithTags("v1.4.0", "stable").Validate())
}

// A tag is a tag, not a reference: the repository is the app's, and the
// environment prefix is added by the build.
func TestDeployRefusesReferencesAndPaths(t *testing.T) {
	for _, tag := range []string{"api:v1", "team/api:v1", "", "-leading", strings.Repeat("x", 101)} {
		assert.NotEmpty(t, reqWithTags(tag).Validate(), "tag %q must be refused", tag)
	}
}

func TestDeployRefusesTooManyTags(t *testing.T) {
	assert.NotEmpty(t, reqWithTags("a", "b", "c", "d", "e", "f").Validate())
}
