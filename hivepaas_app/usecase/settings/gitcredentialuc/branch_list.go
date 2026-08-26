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

func (uc *UC) ListBranch(
	ctx context.Context,
	auth *basedto.Auth,
	req *gitcredentialdto.ListBranchReq,
) (*gitcredentialdto.ListBranchResp, error) {
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, req.Scope, "", req.ID, true,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeGithubApp, base.SettingTypeAccessToken),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	gitResp, err := gitapi.ListBranch(ctx, setting, &gitapi.ListBranchReq{
		Owner:  req.Owner,
		Repo:   req.Repo,
		Paging: &req.Paging,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var branchResp []*gitcredentialdto.BranchResp
	switch gitResp.GitSource {
	case base.GitSourceGithub:
		branchResp, err = gitcredentialdto.TransformGithubBranches(gitResp.GithubBranches)
	case base.GitSourceGitlab:
		branchResp, err = gitcredentialdto.TransformGitlabBranches(gitResp.GitlabBranches)
	case base.GitSourceGitea:
		branchResp, err = gitcredentialdto.TransformGiteaBranches(gitResp.GiteaBranches)
	case base.GitSourceBitbucket, base.GitSourceGogs:
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListBranchResp{
		Meta: &basedto.ListMeta{Page: gitResp.PagingMeta},
		Data: branchResp,
	}, nil
}
