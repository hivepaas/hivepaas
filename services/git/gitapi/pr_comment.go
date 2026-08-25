package gitapi

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

func CreatePullRequestComment(
	ctx context.Context,
	setting *entity.Setting,
	owner string,
	repo string,
	pullNumber int,
	message string,
) error {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		client, err := github.NewFromSetting(setting)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if _, err = client.CreatePullRequestComment(ctx, owner, repo, pullNumber, message); err != nil {
			return apperrors.Wrap(err)
		}

	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) {
		case base.GitSourceGithub:
			client, err := github.NewFromSetting(setting)
			if err != nil {
				return apperrors.Wrap(err)
			}
			if _, err = client.CreatePullRequestComment(ctx, owner, repo, pullNumber, message); err != nil {
				return apperrors.Wrap(err)
			}

		case base.GitSourceGitlab:
			client, err := gitlab.NewFromSetting(setting)
			if err != nil {
				return apperrors.Wrap(err)
			}
			projectID := owner + "/" + repo
			if _, err = client.CreatePullRequestComment(ctx, projectID, pullNumber, message); err != nil {
				return apperrors.Wrap(err)
			}

		case base.GitSourceGitea:
			client, err := gitea.NewFromSetting(setting)
			if err != nil {
				return apperrors.Wrap(err)
			}
			if _, err = client.CreatePullRequestComment(ctx, owner, repo, pullNumber, message); err != nil {
				return apperrors.Wrap(err)
			}

		case base.GitSourceBitbucket, base.GitSourceGogs:
			return nil
		default:
			return nil
		}
	}

	return nil
}
