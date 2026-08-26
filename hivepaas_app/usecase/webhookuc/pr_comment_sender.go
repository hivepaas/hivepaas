package webhookuc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

const (
	prCommentHelpBody = `### 🚀 Deploy Preview
` + "```bash" + `
/hivepaas deploy [subdomain=<name>] [clonedb|noclonedb] [nowait] [nostart]
` + "```" + `
**Available flags:**
- ` + "`subdomain=<name>`" + `: Set a custom subdomain (default: ` + "`pr-<number>`" + `).
- ` + "`clonedb` / `noclonedb`" + `: Force clone or skip cloning configured database apps.
- ` + "`nowait`" + `: Deploy immediately, ignoring the configured creation delay.
- ` + "`nostart`" + `: Create the preview app without starting containers.

### 🛑 Cancel Preview
` + "```bash" + `
/hivepaas cancel
` + "```"

	prCommentDBWarning = "> ⚠️ **Warning:** Database cloning is not enabled for this preview deployment. " +
		"If this pull request introduces database migrations or schema alterations, " +
		"it may directly modify the parent/main application's database.\n\n"
)

func buildInvalidCommandComment(commandText string) string {
	commandText = strings.TrimSpace(commandText)
	return fmt.Sprintf("❌ **Invalid HivePaaS command:** `%s`\n\nHere is the list of available commands:\n\n%s",
		commandText, prCommentHelpBody)
}

func buildDeployPreviewComment(cloneDBApps bool) string {
	var sb strings.Builder
	sb.WriteString("🚀 **HivePaaS is preparing a preview deployment for this pull request...**\n\n")

	if !cloneDBApps {
		sb.WriteString(prCommentDBWarning)
	}

	sb.WriteString("<details>\n<summary>📖 <b>Available commands and options</b></summary>\n\n")
	sb.WriteString(prCommentHelpBody)
	sb.WriteString("\n</details>")

	return sb.String()
}

func buildCancelPreviewComment() string {
	return "🛑 **HivePaaS is canceling and removing preview deployment for this pull request...**"
}

func buildAppNotFoundComment(repoName string) string {
	return fmt.Sprintf("⚠️ **No matching application found in HivePaaS for repository `%s`.**", repoName)
}

func buildPreviewDisabledComment(appName string) string {
	return fmt.Sprintf("⚠️ **Preview deployments are disabled for application `%s`.**\n\n"+
		"Please enable Preview Deployments in the application configuration settings "+
		"on the HivePaaS Dashboard to use this command.", appName)
}

func buildNoActivePreviewComment() string {
	return "ℹ️ **No active preview deployment found for this pull request.**"
}

func buildDeployFailedComment(appName string, err error) string {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	return fmt.Sprintf("❌ **Failed to trigger preview deployment for `%s`:** `%s`", appName, errMsg)
}

// sendPRComment sends a comment back to the Pull Request / Merge Request on GitHub, GitLab, or Gitea.
func (uc *UC) sendPRComment(
	ctx context.Context,
	db database.IDB,
	prCommentEvent *repoPRCommentEventData,
	data *handleRepoWebhookData,
	app *entity.App,
	message string,
) error {
	if prCommentEvent == nil || prCommentEvent.RepoURL == "" || prCommentEvent.PRNumber <= 0 || message == "" {
		return nil
	}

	parsedURL, err := vcsurl.Parse(prCommentEvent.RepoURL)
	if err != nil {
		return apperrors.Wrap(err)
	}

	owner := parsedURL.Username
	repo := parsedURL.Name
	prNumber := int(prCommentEvent.PRNumber)

	// Case 1: Webhook setting is directly a Github App
	if data.WebhookSetting != nil && data.WebhookSetting.Type == base.SettingTypeGithubApp {
		err = gitapi.CreatePullRequestCommentWithRetry(ctx, data.WebhookSetting, owner, repo, prNumber, message)
		return apperrors.Wrap(err)
	}

	// Case 2: Resolve credentials from the app's deployment settings
	if app == nil {
		return nil
	}

	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil
	}

	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return apperrors.Wrap(err)
	}
	if deploymentSettings.RepoSource == nil || deploymentSettings.RepoSource.Credentials.ID == "" {
		return nil
	}

	credSetting, err := uc.settingRepo.GetByID(ctx, db, app.GetObjectScope(), "",
		deploymentSettings.RepoSource.Credentials.ID, true)
	if err != nil {
		return apperrors.Wrap(err)
	}
	if credSetting == nil {
		return nil
	}
	err = gitapi.CreatePullRequestCommentWithRetry(ctx, credSetting, owner, repo, prNumber, message)
	return apperrors.Wrap(err)
}
