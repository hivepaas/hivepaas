package gitapi

import (
	"context"

	gogitea "code.gitea.io/sdk/gitea"
	gogithub "github.com/google/go-github/v85/github"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

type ListRepoReq struct {
	Paging *basedto.Paging
}

type ListRepoResp struct {
	GitSource      base.GitSource
	PagingMeta     *basedto.PagingMeta
	GithubRepos    []*gogithub.Repository
	GitlabProjects []*gogitlab.Project
	GiteaRepos     []*gogitea.Repository
}

func ListRepo(
	ctx context.Context,
	setting *entity.Setting,
	req *ListRepoReq,
) (resp *ListRepoResp, err error) {
	resp = &ListRepoResp{}
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		err = listGithubRepo(ctx, setting, req, resp)

	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) {
		case base.GitSourceGithub:
			err = listGithubRepo(ctx, setting, req, resp)
		case base.GitSourceGitlab:
			err = listGitlabRepo(ctx, setting, req, resp)
		case base.GitSourceGitea:
			err = listGiteaRepo(ctx, setting, req, resp)
		case base.GitSourceBitbucket, base.GitSourceGogs:
			fallthrough
		default:
			return nil, hperrors.Wrap(hperrors.ErrGitTypeUnsupported).WithParam("Type", setting.Kind)
		}

	default:
		return nil, hperrors.Wrap(hperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func listGithubRepo(
	ctx context.Context,
	setting *entity.Setting,
	req *ListRepoReq,
	resp *ListRepoResp,
) (err error) {
	resp.GitSource = base.GitSourceGithub
	client, err := github.NewFromSetting(setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if client.IsAppClient() {
		resp.GithubRepos, resp.PagingMeta, err = client.ListAppRepos(ctx, req.Paging)
	} else {
		resp.GithubRepos, resp.PagingMeta, err = client.ListUserRepos(ctx, req.Paging)
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func listGitlabRepo(
	ctx context.Context,
	setting *entity.Setting,
	req *ListRepoReq,
	resp *ListRepoResp,
) (err error) {
	resp.GitSource = base.GitSourceGitlab
	client, err := gitlab.NewFromSetting(setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	resp.GitlabProjects, resp.PagingMeta, err = client.ListProjects(ctx, req.Paging)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func listGiteaRepo(
	ctx context.Context,
	setting *entity.Setting,
	req *ListRepoReq,
	resp *ListRepoResp,
) (err error) {
	resp.GitSource = base.GitSourceGitea
	client, err := gitea.NewFromSetting(setting)
	if err != nil {
		return hperrors.Wrap(err)
	}
	resp.GiteaRepos, resp.PagingMeta, err = client.ListRepos(ctx, req.Paging)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
