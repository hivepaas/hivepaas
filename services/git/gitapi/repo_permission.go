package gitapi

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

// RepoUser is a user of a git provider: by name on Gitea, by id on GitLab.
type RepoUser struct {
	Login string
	ID    int64
}

// CanWriteRepo answers whether a user may write to a repository, asked of the
// provider with the setting's token. Gitea and GitLab only: GitHub says it in
// its webhooks.
func CanWriteRepo(
	ctx context.Context,
	setting *entity.Setting,
	owner string,
	repo string,
	user RepoUser,
) (bool, error) {
	if setting.Type != base.SettingTypeAccessToken {
		return false, hperrors.Wrap(hperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}
	switch base.GitSource(setting.Kind) { //nolint:exhaustive
	case base.GitSourceGitea:
		if user.Login == "" {
			return false, nil
		}
		client, err := gitea.NewFromSetting(setting)
		if err != nil {
			return false, hperrors.Wrap(err)
		}
		canWrite, err := client.CanWrite(owner, repo, user.Login)
		return canWrite, hperrors.Wrap(err)

	case base.GitSourceGitlab:
		if user.ID == 0 {
			return false, nil
		}
		client, err := gitlab.NewFromSetting(setting)
		if err != nil {
			return false, hperrors.Wrap(err)
		}
		canWrite, err := client.CanWrite(ctx, owner+"/"+repo, user.ID)
		return canWrite, hperrors.Wrap(err)

	default:
		return false, hperrors.Wrap(hperrors.ErrGitTypeUnsupported).WithParam("Type", setting.Kind)
	}
}
