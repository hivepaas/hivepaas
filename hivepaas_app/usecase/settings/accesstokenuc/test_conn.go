package accesstokenuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/hivepaas/hivepaas/services/cloudflare/cfutil"
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
		err = cfutil.VerifyToken(ctx, req.Token)

	default:
		err = hperrors.Wrap(hperrors.ErrTokenTypeUnsupported).WithParam("Type", req.Kind)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &accesstokendto.TestAccessTokenConnResp{}, nil
}
