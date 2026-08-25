package webhookuc

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePRCommentCommand(t *testing.T) {
	uc := &UC{}

	t.Run("Non-hivepaas comment", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "LGTM! Nice work."}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.False(t, isHivepaasCmd)
		assert.False(t, success)
		assert.Empty(t, rawCmd)
	})

	t.Run("Empty /hivepaas comment without subcommand", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.False(t, success)
		assert.Equal(t, "/hivepaas", rawCmd)
	})

	t.Run("Invalid subcommand", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas deployy"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.False(t, success)
		assert.Equal(t, "/hivepaas deployy", rawCmd)
	})

	t.Run("Invalid boolean argument", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas deploy clonedb=invalid"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.Error(t, err)
		assert.True(t, isHivepaasCmd)
		assert.False(t, success)
		assert.Equal(t, "/hivepaas deploy clonedb=invalid", rawCmd)
	})

	t.Run("Unknown flag argument", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas deploy unknownflag=123"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.False(t, success)
		assert.Equal(t, "/hivepaas deploy unknownflag=123", rawCmd)
	})

	t.Run("Valid deploy command default flags", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas deploy"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.True(t, success)
		assert.Equal(t, "/hivepaas deploy", rawCmd)
		assert.Equal(t, previewCmdDeploy, event.previewCmd)
		assert.False(t, event.previewDeployNoStart)
		assert.False(t, event.previewDeployNoWait)
		assert.False(t, event.previewDeployCloneDB)
		assert.False(t, event.previewDeployNoCloneDB)
		assert.Empty(t, event.previewDeploySubdomain)
	})

	t.Run("Valid deploy command with multiple flags", func(t *testing.T) {
		event := &repoPRCommentEventData{
			CommentBody: "\n\n  /hivepaas deploy subdomain=feature-abc clonedb nowait nostart\nSome extra notes",
		}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.True(t, success)
		assert.Equal(t, "/hivepaas deploy subdomain=feature-abc clonedb nowait nostart", rawCmd)
		assert.Equal(t, previewCmdDeploy, event.previewCmd)
		assert.Equal(t, "feature-abc", event.previewDeploySubdomain)
		assert.True(t, event.previewDeployCloneDB)
		assert.True(t, event.previewDeployNoWait)
		assert.True(t, event.previewDeployNoStart)
	})

	t.Run("Valid deploy command with noclonedb", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas deploy noclonedb"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.True(t, success)
		assert.Equal(t, "/hivepaas deploy noclonedb", rawCmd)
		assert.Equal(t, previewCmdDeploy, event.previewCmd)
		assert.True(t, event.previewDeployNoCloneDB)
	})

	t.Run("Valid cancel command", func(t *testing.T) {
		event := &repoPRCommentEventData{CommentBody: "/hivepaas cancel"}
		isHivepaasCmd, success, rawCmd, err := uc.parsePRCommentCommand(event)
		assert.NoError(t, err)
		assert.True(t, isHivepaasCmd)
		assert.True(t, success)
		assert.Equal(t, "/hivepaas cancel", rawCmd)
		assert.Equal(t, previewCmdCancel, event.previewCmd)
	})
}

func TestBuildInvalidCommandComment(t *testing.T) {
	comment := buildInvalidCommandComment("/hivepaas deployy")
	assert.Contains(t, comment, "❌ **Invalid HivePaaS command:** `/hivepaas deployy`")
	assert.Contains(t, comment, "/hivepaas deploy [subdomain=<name>] [clonedb|noclonedb] [nowait] [nostart]")
	assert.Contains(t, comment, "/hivepaas cancel")
}

func TestBuildDeployPreviewComment(t *testing.T) {
	t.Run("Without clone DB apps (shows migration warning)", func(t *testing.T) {
		comment := buildDeployPreviewComment(false)
		assert.Contains(t, comment, "🚀 **HivePaaS is preparing a preview deployment for this pull request...**")
		assert.Contains(t, comment, "> ⚠️ **Warning:** Database cloning is not enabled for this preview deployment")
		assert.Contains(t, comment, "<details>")
		assert.Contains(t, comment, "<summary>📖 <b>Available commands and options</b></summary>")
		assert.Contains(t, comment, "</details>")
	})

	t.Run("With clone DB apps (hides migration warning)", func(t *testing.T) {
		comment := buildDeployPreviewComment(true)
		assert.Contains(t, comment, "🚀 **HivePaaS is preparing a preview deployment for this pull request...**")
		assert.False(t, strings.Contains(comment, "> ⚠️ **Warning:** Database cloning is not enabled"))
		assert.Contains(t, comment, "<details>")
		assert.Contains(t, comment, "<summary>📖 <b>Available commands and options</b></summary>")
	})
}

func TestBuildCancelPreviewComment(t *testing.T) {
	comment := buildCancelPreviewComment()
	assert.Contains(t, comment, "🛑 **HivePaaS is canceling and removing preview deployment for this pull request...**")
}

func TestBuildErrorComments(t *testing.T) {
	t.Run("App Not Found", func(t *testing.T) {
		comment := buildAppNotFoundComment("my-backend-repo")
		assert.Contains(t, comment, "⚠️ **No matching application found in HivePaaS for repository `my-backend-repo`.**")
	})

	t.Run("Preview Disabled", func(t *testing.T) {
		comment := buildPreviewDisabledComment("web-app")
		assert.Contains(t, comment, "⚠️ **Preview deployments are disabled for application `web-app`.**")
	})

	t.Run("No Active Preview to cancel", func(t *testing.T) {
		comment := buildNoActivePreviewComment()
		assert.Contains(t, comment, "ℹ️ **No active preview deployment found for this pull request.**")
	})

	t.Run("Deploy Failed", func(t *testing.T) {
		comment := buildDeployFailedComment("web-app", errors.New("database connection timeout"))
		assert.Contains(t, comment, "❌ **Failed to trigger preview deployment for `web-app`:**")
		assert.Contains(t, comment, "database connection timeout")
	})
}
