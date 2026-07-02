package githubappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/services/git/github"
)

func (uc *UC) ListAppInstallation(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.ListAppInstallationReq,
) (*githubappdto.ListAppInstallationResp, error) {
	client, err := github.NewFromApp(req.GhAppID, req.GhInstallationID, reflectutil.UnsafeStrToBytes(req.PrivateKey))
	if err != nil {
		return nil, apperrors.New(err)
	}

	installations, pagingMeta, err := client.ListInstallations(ctx, &req.Paging)
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := githubappdto.TransformAppInstallations(installations)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &githubappdto.ListAppInstallationResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
