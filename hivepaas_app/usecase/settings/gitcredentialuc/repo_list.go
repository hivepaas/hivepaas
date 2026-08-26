package gitcredentialuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

func (uc *UC) ListRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *gitcredentialdto.ListRepoReq,
) (resp *gitcredentialdto.ListRepoResp, err error) {
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, req.Scope, "", req.ID, true,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeGithubApp, base.SettingTypeAccessToken),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	gitResp, err := gitapi.ListRepo(ctx, setting, &gitapi.ListRepoReq{
		Paging: &req.Paging,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var repoResp []*gitcredentialdto.RepoResp
	switch gitResp.GitSource {
	case base.GitSourceGithub:
		repoResp, err = gitcredentialdto.TransformGithubRepos(gitResp.GithubRepos)
	case base.GitSourceGitlab:
		repoResp, err = gitcredentialdto.TransformGitlabProjects(gitResp.GitlabProjects)
	case base.GitSourceGitea:
		repoResp, err = gitcredentialdto.TransformGiteaRepos(gitResp.GiteaRepos)
	case base.GitSourceBitbucket, base.GitSourceGogs:
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListRepoResp{
		Meta: &basedto.ListMeta{Page: gitResp.PagingMeta},
		Data: repoResp,
	}, nil
}
