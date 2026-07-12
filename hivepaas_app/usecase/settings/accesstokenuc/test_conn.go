package accesstokenuc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httputil"
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
	case base.AccessTokenKindCloudflare:
		err = uc.testCloudflareTokenValid(ctx, req)
	case base.AccessTokenKindBitbucket, base.AccessTokenKindGogs:
		fallthrough
	default:
		err = apperrors.Wrap(apperrors.ErrTokenTypeUnsupported).WithParam("Type", req.Kind)
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
	client, err := github.NewFromToken(req.Token)
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

func (uc *UC) testCloudflareTokenValid(
	ctx context.Context,
	req *accesstokendto.TestAccessTokenConnReq,
) error {
	data, err := httputil.HTTPGet(ctx, "https://api.cloudflare.com/client/v4/user/tokens/verify",
		func(httpReq *http.Request) {
			httpReq.Header.Set("Authorization", "Bearer "+req.Token)
			httpReq.Header.Set("Content-Type", "application/json")
		})
	if err != nil {
		return apperrors.Wrap(apperrors.ErrTokenInvalid).WithCause(err)
	}

	var cloudflareResp struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(data, &cloudflareResp); err != nil {
		return apperrors.Wrap(err)
	}

	if !cloudflareResp.Success {
		return apperrors.Wrap(apperrors.ErrTokenInvalid).WithMsgLog(
			"Cloudflare token verification response success was false")
	}

	return nil
}
