package githubappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

func (uc *UC) TestGithubAppConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.TestGithubAppConnReq,
) (*githubappdto.TestGithubAppConnResp, error) {
	client, err := github.NewFromApp(req.GhAppID, req.GhInstallationID, reflectutil.UnsafeStrToBytes(req.PrivateKey))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	_, _, err = client.ListInstallations(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.TestGithubAppConnResp{}, nil
}
