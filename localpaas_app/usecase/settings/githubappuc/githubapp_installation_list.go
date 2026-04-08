package githubappuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/services/git/github"
)

func (uc *UC) ListAppInstallation(
	ctx context.Context,
	auth *basedto.Auth,
	req *githubappdto.ListAppInstallationReq,
) (*githubappdto.ListAppInstallationResp, error) {
	client, err := github.NewFromApp(req.GhAppID, req.GhInstallationID, reflectutil.UnsafeStrToBytes(req.PrivateKey))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	installations, pagingMeta, err := client.ListInstallations(ctx, &req.Paging)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp, err := githubappdto.TransformAppInstallations(installations)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &githubappdto.ListAppInstallationResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
