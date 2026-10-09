package appdeploymentserviceimpl

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestFailureReason(t *testing.T) {
	assert.Empty(t, failureReason(&entity.Deployment{}), "no output, no reason")

	deployment := &entity.Deployment{Output: &entity.AppDeploymentOutput{
		Errors: []string{"[WARN] no registry to push to", "image not found"},
	}}
	assert.Equal(t, "[WARN] no registry to push to\nimage not found", failureReason(deployment))

	// A long one is cut to what a notification field takes, whole characters.
	deployment.Output.Errors = []string{strings.Repeat("é", 3*notifReasonMaxLen)}
	reason := failureReason(deployment)
	assert.Equal(t, notifReasonMaxLen, utf8.RuneCountInString(reason))
	assert.True(t, utf8.ValidString(reason))
	assert.True(t, strings.HasSuffix(reason, "…"))
}
