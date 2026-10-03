package webhookuc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
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

// buildDeployPreviewComment answers a deploy command, with the warnings that go
// with it: the database shared with the app, and withheldNote, the secrets the
// preview goes without.
func buildDeployPreviewComment(cloneDBApps bool, withheldNote string) string {
	var sb strings.Builder
	sb.WriteString("🚀 **HivePaaS is preparing a preview deployment for this pull request...**\n\n")

	if !cloneDBApps {
		sb.WriteString(prCommentDBWarning)
	}
	sb.WriteString(withheldNote)

	sb.WriteString("<details>\n<summary>📖 <b>Available commands and options</b></summary>\n\n")
	sb.WriteString(prCommentHelpBody)
	sb.WriteString("\n</details>")

	return sb.String()
}

// buildWithheldSecretsNote warns that the preview goes without the app's secrets
// that are not inheritable, and empties the variables using them. It names the
// secrets and the variables - never a value - and says where to change that.
// "" when there are none.
func buildWithheldSecretsNote(appName, secretsURL string, secrets []*apppreviewservice.WithheldSecret) string {
	if len(secrets) == 0 {
		return ""
	}
	where := "**Secrets**"
	if secretsURL != "" {
		where = "[**Secrets**](" + secretsURL + ")"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "> ⚠️ **Warning:** These secrets of `%s` are not inheritable, so the preview does not get "+
		"them, and the variables using them are empty in it:\n>\n", appName)
	for _, secret := range secrets {
		fmt.Fprintf(&sb, "> - `%s`, used by `%s`\n", secret.Name, strings.Join(secret.EnvVars, "`, `"))
	}
	fmt.Fprintf(&sb, ">\n> To give the preview a secret, turn on **Inheritable** for it in the application's %s "+
		"on the HivePaaS Dashboard, then deploy the preview again.\n\n", where)
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

// buildPRCommentsDisabledComment refuses a comment's command on an app whose
// previews are not run from comments, saying where that is turned on.
func buildPRCommentsDisabledComment(appName, settingsURL string) string {
	where := "**Feature Settings**"
	if settingsURL != "" {
		where = "[**Feature Settings**](" + settingsURL + ")"
	}
	return fmt.Sprintf("⚠️ **Pull request comments cannot deploy or cancel previews of application `%s`.**\n\n"+
		"To use `/hivepaas deploy` and `/hivepaas cancel`, turn on **Allow PR Comments** under App Preview, "+
		"in the application's %s on the HivePaaS Dashboard.", appName, where)
}

// buildAuthorNotAllowedComment refuses a command from someone who may not write
// to the repository.
func buildAuthorNotAllowedComment() string {
	return "⚠️ **Only people who can write to this repository can run HivePaaS commands.**\n\n" +
		"A preview runs the pull request's code with the application's environment variables, so it is started " +
		"by the repository's owners, members and collaborators only."
}

// buildPushNotDeployedComment says a stranger's new commits were not deployed
// to the pull request's preview, and how someone who may write deploys them.
func buildPushNotDeployedComment(changeID string) string {
	commit := "New commits were"
	if changeID != "" {
		commit = fmt.Sprintf("New commits, up to `%.12s`, were", changeID)
	}
	return "⚠️ **" + commit + " pushed, but the preview was not deployed again.**\n\n" +
		"The pull request's author cannot write to this repository, and a preview runs its code with the " +
		"application's environment variables. Someone who can should read the changes, then comment " +
		"`/hivepaas deploy` to deploy them."
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
		return hperrors.Wrap(err)
	}

	owner := parsedURL.Username
	repo := parsedURL.Name
	prNumber := int(prCommentEvent.PRNumber)

	setting, err := uc.gitAPISetting(ctx, db, data, app)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		return nil
	}
	err = gitapi.CreatePullRequestCommentWithRetry(ctx, setting, owner, repo, prNumber, message)
	return hperrors.Wrap(err)
}
