package gitcredentialuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
	"github.com/localpaas/localpaas/services/git/gitea"
	"github.com/localpaas/localpaas/services/git/github"
	"github.com/localpaas/localpaas/services/git/gitlab"
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
		return nil, apperrors.New(err)
	}

	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		return uc.listGithubBranch(ctx, req, setting)
	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) {
		case base.GitSourceGithub:
			return uc.listGithubBranch(ctx, req, setting)
		case base.GitSourceGitlab:
			return uc.listGitlabBranch(ctx, req, setting)
		case base.GitSourceGitea:
			return uc.listGiteaBranch(ctx, req, setting)
		case base.GitSourceBitbucket, base.GitSourceGogs:
			fallthrough
		default:
			return nil, apperrors.New(apperrors.ErrGitTypeUnsupported).WithParam("Type", setting.Kind)
		}
	default:
		return nil, apperrors.New(apperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}
}

func (uc *UC) listGithubBranch(
	ctx context.Context,
	req *gitcredentialdto.ListBranchReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListBranchResp, error) {
	client, err := github.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// If setting is a github-app, we get `owner` from the setting
	if setting.Type == base.SettingTypeGithubApp {
		githubApp := setting.MustAsGithubApp()
		if githubApp.Organization != "" && req.Owner != "" && githubApp.Organization != req.Owner {
			return nil, apperrors.NewMismatch("owner", "organization")
		}
		req.Owner = gofn.Coalesce(req.Owner, githubApp.Organization)
	}

	branches, pagingMeta, err := client.ListBranch(ctx, req.Owner, req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGithubBranches(branches)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListBranchResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGitlabBranch(
	ctx context.Context,
	req *gitcredentialdto.ListBranchReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListBranchResp, error) {
	client, err := gitlab.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	branches, pagingMeta, err := client.ListBranch(ctx, req.Owner+"/"+req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGitlabBranches(branches)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListBranchResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGiteaBranch(
	ctx context.Context,
	req *gitcredentialdto.ListBranchReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListBranchResp, error) {
	client, err := gitea.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	branches, pagingMeta, err := client.ListBranch(ctx, req.Owner, req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGiteaBranches(branches)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListBranchResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
