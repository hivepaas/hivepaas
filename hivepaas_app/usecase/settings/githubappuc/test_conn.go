package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/services/git/github"
)

func (uc *UC) TestGithubAppConn(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.TestGithubAppConnReq,
) (*githubappdto.TestGithubAppConnResp, error) {
	client, err := github.NewFromApp(req.GhAppID, req.GhInstallationID, reflectutil.UnsafeStrToBytes(req.PrivateKey))
	if err != nil {
		return nil, apperrors.New(err)
	}

	_, _, err = client.ListInstallations(ctx, &basedto.Paging{Limit: 1})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.TestGithubAppConnResp{}, nil
}
