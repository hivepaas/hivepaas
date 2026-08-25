package appdeploymentserviceimpl

import (
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

func TestBuildDeploySucceededPRComment(t *testing.T) {
	comment := buildDeploySucceededPRComment(
		"my-preview-app",
		"https://pr-123.example.com",
		"`a1b2c3d` - *Feature X* (@dev)",
		"1m 24s",
		"https://dashboard.example.com/deployments/dep-1",
	)

	assert.Contains(t, comment, "### 🚀 **Preview Deployment Ready!**")
	assert.Contains(t, comment, "| **Preview URL** | 🔗 [https://pr-123.example.com](https://pr-123.example.com) |")
	assert.Contains(t, comment, "| **Application** | `my-preview-app` |")
	assert.Contains(t, comment, "| **Commit** | `a1b2c3d` - *Feature X* (@dev) |")
	assert.Contains(t, comment, "| **Duration** | ⏱️ 1m 24s |")
	assert.Contains(t, comment, "👉 [**View Deployment Details on HivePaaS**]")
}

func TestBuildDeployFailedPRComment(t *testing.T) {
	comment := buildDeployFailedPRComment(
		"my-preview-app",
		"Dockerfile line 12: yarn build failed with code 1",
		"`a1b2c3d` - *Feature X* (@dev)",
		"45s",
		"https://dashboard.example.com/deployments/dep-1",
	)

	assert.Contains(t, comment, "### ❌ **Preview Deployment Failed**")
	assert.Contains(t, comment, "The deployment for application `my-preview-app` encountered an error:")
	assert.Contains(t, comment, "Dockerfile line 12: yarn build failed with code 1")
	assert.Contains(t, comment, "| **Commit** | `a1b2c3d` - *Feature X* (@dev) |")
	assert.Contains(t, comment, "| **Duration** | ⏱️ 45s (Failed) |")
	assert.Contains(t, comment, "🔍 [**View Full Build Logs on HivePaaS Dashboard**]")
}
