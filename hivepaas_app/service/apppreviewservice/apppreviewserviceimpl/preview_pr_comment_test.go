package apppreviewserviceimpl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestExtractPRNumber(t *testing.T) {
	t.Run("Github pull ref", func(t *testing.T) {
		prNum := extractPRNumber("pull/123", nil)
		assert.Equal(t, 123, prNum)
	})

	t.Run("Gitlab merge request ref", func(t *testing.T) {
		prNum := extractPRNumber("merge-requests/456", nil)
		assert.Equal(t, 456, prNum)
	})

	t.Run("Full refs pull head", func(t *testing.T) {
		prNum := extractPRNumber("refs/pull/789/head", nil)
		assert.Equal(t, 789, prNum)
	})

	t.Run("From Trigger ChangeID", func(t *testing.T) {
		trigger := &entity.AppDeploymentTrigger{
			Source:   base.DeploymentTriggerSourceRepoWebhook,
			ChangeID: "pr-999",
		}
		prNum := extractPRNumber("main", trigger)
		assert.Equal(t, 999, prNum)
	})

	t.Run("Non PR ref and no trigger", func(t *testing.T) {
		prNum := extractPRNumber("main", nil)
		assert.Equal(t, 0, prNum)
	})
}

func TestBuildPreviewFailedPRComment(t *testing.T) {
	err := errors.New("failed to clone database app 'postgres-main': connection refused")
	comment := buildPreviewFailedPRComment("web-app", "pull/42", err)

	assert.Contains(t, comment, "### ❌ **Preview Environment Creation Failed**")
	assert.Contains(t, comment, "Failed to initialize preview environment for application `web-app`:")
	assert.Contains(t, comment, "failed to clone database app 'postgres-main': connection refused")
	assert.Contains(t, comment, "| **Application** | `web-app` |")
	assert.Contains(t, comment, "| **Target Ref** | `pull/42` |")
	assert.Contains(t, comment, "👉 *Please check your application preview settings and database dependencies")
}
