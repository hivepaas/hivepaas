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

func (uc *UC) ListPullRequest(
	ctx context.Context,
	auth *basedto.Auth,
	req *gitcredentialdto.ListPullRequestReq,
) (*gitcredentialdto.ListPullRequestResp, error) {
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, req.Scope, "", req.ID, true,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeGithubApp, base.SettingTypeAccessToken),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	gitResp, err := gitapi.ListPullRequest(ctx, setting, &gitapi.ListPullRequestReq{
		Owner:  req.Owner,
		Repo:   req.Repo,
		Paging: &req.Paging,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var prResp []*gitcredentialdto.PullRequestResp
	switch gitResp.GitSource {
	case base.GitSourceGithub:
		prResp, err = gitcredentialdto.TransformGithubPullRequests(gitResp.GithubPullRequests)
	case base.GitSourceGitlab:
		prResp, err = gitcredentialdto.TransformGitlabMergeRequests(gitResp.GitlabMergeRequests)
	case base.GitSourceGitea:
		prResp, err = gitcredentialdto.TransformGiteaPullRequests(gitResp.GiteaPullRequests)
	case base.GitSourceBitbucket, base.GitSourceGogs:
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListPullRequestResp{
		Meta: &basedto.ListMeta{Page: gitResp.PagingMeta},
		Data: prResp,
	}, nil
}
