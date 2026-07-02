package accesstokenuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/hivepaas/hivepaas/services/git/gitea"
	"github.com/hivepaas/hivepaas/services/git/github"
	"github.com/hivepaas/hivepaas/services/git/gitlab"
)

func (uc *UC) TestAccessTokenConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.TestAccessTokenConnReq,
) (*accesstokendto.TestAccessTokenConnResp, error) {
	var err error
	switch req.Kind {
	case base.AccessTokenKindGithub:
		err = uc.testGithubTokenConn(ctx, req)
	case base.AccessTokenKindGitlab:
		err = uc.testGitlabTokenConn(ctx, req)
	case base.AccessTokenKindGitea:
		err = uc.testGiteaTokenConn(ctx, req)
	case base.AccessTokenKindBitbucket, base.AccessTokenKindGogs:
		fallthrough
	default:
		err = apperrors.New(apperrors.ErrGitTypeUnsupported).WithParam("Type", req.Kind)
	}
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &accesstokendto.TestAccessTokenConnResp{}, nil
}

func (uc *UC) testGithubTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := github.NewFromPersonalToken(req.Token)
	if err != nil {
		return apperrors.New(err)
	}
	_, _, err = client.ListUserRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) testGitlabTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := gitlab.NewFromToken(req.Token, req.BaseURL)
	if err != nil {
		return apperrors.New(err)
	}
	_, _, err = client.ListAllProjects(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}

func (uc *UC) testGiteaTokenConn(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	client, err := gitea.NewFromToken(req.Token, req.BaseURL)
	if err != nil {
		return apperrors.New(err)
	}
	_, _, err = client.ListAllRepos(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
