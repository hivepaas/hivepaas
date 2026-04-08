package accesstokenuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/localpaas/localpaas/services/git/gitea"
	"github.com/localpaas/localpaas/services/git/github"
	"github.com/localpaas/localpaas/services/git/gitlab"
)

func (uc *UC) TestAccessTokenConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.TestAccessTokenConnReq,
) (*accesstokendto.TestAccessTokenConnResp, error) {
	var err error
	switch req.Kind { //nolint:exhaustive
	case base.AccessTokenKindGithub:
		err = uc.testGithubTokenConn(ctx, req)
	case base.AccessTokenKindGitlab:
		err = uc.testGitlabTokenConn(ctx, req)
	case base.AccessTokenKindGitea:
		err = uc.testGiteaTokenConn(ctx, req)
	default:
		err = apperrors.New(apperrors.ErrUnsupported).
			WithMsgLog("Git source '%s' unsupported", req.Kind)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &accesstokendto.TestAccessTokenConnResp{}, nil
}

func (uc *UC) testGithubTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := github.NewFromPersonalToken(req.Token)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, _, err = client.ListUserRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) testGitlabTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := gitlab.NewFromToken(req.Token, req.BaseURL)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, _, err = client.ListAllProjects(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) testGiteaTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := gitea.NewFromToken(req.Token, req.BaseURL)
	if err != nil {
		return apperrors.Wrap(err)
	}
	_, _, err = client.ListAllRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
