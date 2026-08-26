package gitapi

import (
	"context"

	gogitea "code.gitea.io/sdk/gitea"
	gogithub "github.com/google/go-github/v85/github"
	"github.com/tiendc/gofn"
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

type ListPullRequestReq struct {
	Owner  string
	Repo   string
	Paging *basedto.Paging
}

type ListPullRequestResp struct {
	GitSource           base.GitSource
	PagingMeta          *basedto.PagingMeta
	GithubPullRequests  []*gogithub.PullRequest
	GitlabMergeRequests []*gogitlab.BasicMergeRequest
	GiteaPullRequests   []*gogitea.PullRequest
}

func ListPullRequest(
	ctx context.Context,
	setting *entity.Setting,
	req *ListPullRequestReq,
) (resp *ListPullRequestResp, err error) {
	resp = &ListPullRequestResp{}
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		err = listGithubPullRequest(ctx, setting, req, resp)

	case base.SettingTypeAccessToken:
		switch base.GitSource(setting.Kind) {
		case base.GitSourceGithub:
			err = listGithubPullRequest(ctx, setting, req, resp)
		case base.GitSourceGitlab:
			err = listGitlabPullRequest(ctx, setting, req, resp)
		case base.GitSourceGitea:
			err = listGiteaPullRequest(ctx, setting, req, resp)
		case base.GitSourceBitbucket, base.GitSourceGogs:
			fallthrough
		default:
			return nil, apperrors.Wrap(apperrors.ErrGitTypeUnsupported).WithParam("Type", setting.Kind)
		}

	default:
		return nil, apperrors.Wrap(apperrors.ErrSettingTypeUnsupported).WithParam("Name", setting.Type)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func listGithubPullRequest(
	ctx context.Context,
	setting *entity.Setting,
	req *ListPullRequestReq,
	resp *ListPullRequestResp,
) (err error) {
	resp.GitSource = base.GitSourceGithub
	client, err := github.NewFromSetting(setting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If setting is a github-app, we get `owner` from the setting
	owner := req.Owner
	if setting.Type == base.SettingTypeGithubApp {
		githubApp := setting.MustAsGithubApp()
		if githubApp.Organization != "" && req.Owner != "" && githubApp.Organization != req.Owner {
			return apperrors.NewMismatch("owner", "organization")
		}
		owner = gofn.Coalesce(owner, githubApp.Organization)
	}

	resp.GithubPullRequests, resp.PagingMeta, err = client.ListPullRequest(ctx, owner, req.Repo, req.Paging)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func listGitlabPullRequest(
	ctx context.Context,
	setting *entity.Setting,
	req *ListPullRequestReq,
	resp *ListPullRequestResp,
) (err error) {
	resp.GitSource = base.GitSourceGitlab
	client, err := gitlab.NewFromSetting(setting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	resp.GitlabMergeRequests, resp.PagingMeta, err = client.ListPullRequest(ctx, req.Owner+"/"+req.Repo, req.Paging)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func listGiteaPullRequest(
	ctx context.Context,
	setting *entity.Setting,
	req *ListPullRequestReq,
	resp *ListPullRequestResp,
) (err error) {
	resp.GitSource = base.GitSourceGitea
	client, err := gitea.NewFromSetting(setting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	resp.GiteaPullRequests, resp.PagingMeta, err = client.ListPullRequest(ctx, req.Owner, req.Repo, req.Paging)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
