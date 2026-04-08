package gitcredentialuc

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v79/github"

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

func (uc *UC) ListRepo(
	ctx context.Context,
	auth *basedto.Auth,
	req *gitcredentialdto.ListRepoReq,
) (*gitcredentialdto.ListRepoResp, error) {
	setting, err := uc.SettingRepo.GetByID(ctx, uc.DB, req.Scope, "", req.ID, true,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeGithubApp, base.SettingTypeAccessToken),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		return uc.listGithubRepo(ctx, req, setting)
	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) { //nolint:exhaustive
		case base.GitSourceGithub:
			return uc.listGithubRepo(ctx, req, setting)
		case base.GitSourceGitlab:
			return uc.listGitlabRepo(ctx, req, setting)
		case base.GitSourceGitea:
			return uc.listGiteaRepo(ctx, req, setting)
		default:
			return nil, apperrors.NewUnsupported(fmt.Sprintf("Git source '%v'", setting.Kind))
		}
	default:
		return nil, apperrors.NewUnsupported(fmt.Sprintf("Setting type '%v'", setting.Type))
	}
}

func (uc *UC) listGithubRepo(
	ctx context.Context,
	req *gitcredentialdto.ListRepoReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListRepoResp, error) {
	client, err := github.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var repos []*gogithub.Repository
	var pagingMeta *basedto.PagingMeta
	if client.IsAppClient() {
		repos, pagingMeta, err = client.ListAppRepos(ctx, &req.Paging)
	} else {
		repos, pagingMeta, err = client.ListUserRepos(ctx, &req.Paging)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := gitcredentialdto.TransformGithubRepos(repos)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListRepoResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGitlabRepo(
	ctx context.Context,
	req *gitcredentialdto.ListRepoReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListRepoResp, error) {
	client, err := gitlab.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	projects, pagingMeta, err := client.ListProjects(ctx, &req.Paging)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := gitcredentialdto.TransformGitlabProjects(projects)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListRepoResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) listGiteaRepo(
	ctx context.Context,
	req *gitcredentialdto.ListRepoReq,
	setting *entity.Setting,
) (*gitcredentialdto.ListRepoResp, error) {
	client, err := gitea.NewFromSetting(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	repos, pagingMeta, err := client.ListRepos(ctx, &req.Paging)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := gitcredentialdto.TransformGiteaRepos(repos)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &gitcredentialdto.ListRepoResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
