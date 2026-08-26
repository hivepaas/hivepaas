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
	"github.com/hivepaas/hivepaas/services/git/gitapi"
)

func (uc *UC) TestAccessTokenConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *accesstokendto.TestAccessTokenConnReq,
) (*accesstokendto.TestAccessTokenConnResp, error) {
	var err error
	switch req.Kind {
	case base.AccessTokenKindGithub, base.AccessTokenKindGitlab, base.AccessTokenKindGitea,
		base.AccessTokenKindBitbucket, base.AccessTokenKindGogs:
		err = gitapi.TestAccessTokenConn(ctx, req.Kind, req.Token, req.BaseURL)

	case base.AccessTokenKindCloudflare:
		err = uc.testCloudflareTokenValid(ctx, req)

	default:
		err = apperrors.Wrap(apperrors.ErrTokenTypeUnsupported).WithParam("Type", req.Kind)
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &accesstokendto.TestAccessTokenConnResp{}, nil
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
