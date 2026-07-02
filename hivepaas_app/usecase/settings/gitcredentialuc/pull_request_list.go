package gitcredentialuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
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
		return nil, apperrors.New(err)
	}

	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		return uc.listGithubPullRequest(ctx, req, setting)
	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) {
		case base.GitSourceGithub:
			return uc.listGithubPullRequest(ctx, req, setting)
		case base.GitSourceGitlab:
			return uc.listGitlabPullRequest(ctx, req, setting)
		case base.GitSourceGitea:
			return uc.listGiteaPullRequest(ctx, req, setting)
		case base.GitSourceBitbucket, base.GitSourceGogs:
			fallthrough
		default:
			return nil, apperrors.New(apperrors.ErrGitTypeUnsupported).WithParam("Type", setting.Kind)
		}
	default:
		return nil, apperrors.New(apperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}
}

func (uc *UC) listGithubPullRequest(
	ctx context.Context,
	req *gitcredentialdto.ListPullRequestReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListPullRequestResp, error) {
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

	prs, pagingMeta, err := client.ListPullRequest(ctx, req.Owner, req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGithubPullRequests(prs)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListPullRequestResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGitlabPullRequest(
	ctx context.Context,
	req *gitcredentialdto.ListPullRequestReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListPullRequestResp, error) {
	client, err := gitlab.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	mrs, pagingMeta, err := client.ListPullRequest(ctx, req.Owner+"/"+req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGitlabMergeRequests(mrs)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListPullRequestResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGiteaPullRequest(
	ctx context.Context,
	req *gitcredentialdto.ListPullRequestReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListPullRequestResp, error) {
	client, err := gitea.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	prs, pagingMeta, err := client.ListPullRequest(ctx, req.Owner, req.Repo, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := gitcredentialdto.TransformGiteaPullRequests(prs)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &gitcredentialdto.ListPullRequestResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
