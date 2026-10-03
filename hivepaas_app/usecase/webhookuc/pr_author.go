package webhookuc

import (
	"context"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

// githubWriterAssociations are the authors GitHub relates to a repository as
// able to write to it: its owner, its organization's members, its
// collaborators. Anyone else - a first-time contributor, a stranger - is not.
var githubWriterAssociations = []string{"OWNER", "MEMBER", "COLLABORATOR"}

// prAuthor is who wrote a pull request, or a comment on one, and what the
// provider's webhook says of them.
type prAuthor struct {
	// Login names the author; ID is GitLab's id of them.
	Login string
	ID    int64
	// Association is GitHub's author_association, how the author is related to
	// the repository; GitHub always sends it, NONE for a stranger.
	Association string
	// IsRepoOwner is true when the author owns the repository.
	IsRepoOwner bool
	// RepoPrivate is true when only those given access to the repository can
	// comment on it.
	RepoPrivate bool
}

// prAuthorAllowed decides whether an author - of a comment's command, or of a
// pull request whose preview new commits would deploy - may have a preview run:
// someone who may write to the repository. A preview runs the pull request's
// code with the app's env vars, so a stranger on a public repository must not
// start one.
func (uc *UC) prAuthorAllowed(
	ctx context.Context,
	db database.IDB,
	repoURL string,
	author *prAuthor,
	data *handleRepoWebhookData,
	apps []*entity.App,
) bool {
	var canWrite *bool
	if prAuthorNeedsAsking(author, data.WebhookSetting.MustAsRepoWebhook().Kind) {
		canWrite = uc.askCanWriteRepo(ctx, db, repoURL, author, data, apps)
	}
	return prAuthorVerdict(author, canWrite)
}

// prAuthorNeedsAsking is true when the provider's API must say whether
// the author may write: Gitea and GitLab do not say it in their webhooks.
func prAuthorNeedsAsking(author *prAuthor, kind base.WebhookKind) bool {
	if author.Association != "" || author.IsRepoOwner {
		return false
	}
	return kind == base.WebhookKindGitea || kind == base.WebhookKindGitlab
}

// prAuthorVerdict is whether the author may run commands, from the
// webhook and, when asked, the provider's answer - canWrite, nil when it was
// not asked or did not answer. With neither, only a private repository's
// commenters may, as only those given access to it can comment.
func prAuthorVerdict(author *prAuthor, canWrite *bool) bool {
	switch {
	case author.Association != "":
		return slices.Contains(githubWriterAssociations, author.Association)
	case author.IsRepoOwner:
		return true
	case canWrite != nil:
		return *canWrite
	}
	return author.RepoPrivate
}

// askCanWriteRepo asks the provider whether the author may write to the
// repository, with the token of the first app that has one. nil when none
// could answer.
func (uc *UC) askCanWriteRepo(
	ctx context.Context,
	db database.IDB,
	repoURL string,
	author *prAuthor,
	data *handleRepoWebhookData,
	apps []*entity.App,
) *bool {
	parsedURL, err := vcsurl.Parse(repoURL)
	if err != nil {
		return nil
	}
	user := gitapi.RepoUser{Login: author.Login, ID: author.ID}
	for _, app := range apps {
		setting, err := uc.gitAPISetting(ctx, db, data, app)
		if err != nil || setting == nil {
			continue
		}
		canWrite, err := gitapi.CanWriteRepo(ctx, setting, parsedURL.Username, parsedURL.Name, user)
		if err != nil {
			logging.Warnf("webhook: asking whether %s may write to %s: %v", user.Login, repoURL, err)
			continue
		}
		return &canWrite
	}
	return nil
}

// gitAPISetting is the setting HivePaaS calls the provider's API with for an
// app's repository: the GitHub App the webhook is of, or the credentials the
// app's deployment pulls with. nil when there is none.
func (uc *UC) gitAPISetting(
	ctx context.Context,
	db database.IDB,
	data *handleRepoWebhookData,
	app *entity.App,
) (*entity.Setting, error) {
	if data.WebhookSetting != nil && data.WebhookSetting.Type == base.SettingTypeGithubApp {
		return data.WebhookSetting, nil
	}
	if app == nil {
		return nil, nil
	}
	deploymentSetting := app.GetSettingByType(base.SettingTypeAppDeployment)
	if deploymentSetting == nil {
		return nil, nil
	}
	deploymentSettings, err := deploymentSetting.AsAppDeploymentSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if deploymentSettings.RepoSource == nil || deploymentSettings.RepoSource.Credentials.ID == "" {
		return nil, nil
	}
	credSetting, err := uc.settingRepo.GetByID(ctx, db, app.GetObjectScope(), "",
		deploymentSettings.RepoSource.Credentials.ID, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return credSetting, nil
}
