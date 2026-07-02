package acmednsprovideruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
	"github.com/hivepaas/hivepaas/services/ssl/acme"
)

func (uc *UC) TestProviderAccess(
	ctx context.Context,
	auth *basedto.Auth,
	req *acmednsproviderdto.TestProviderAccessReq,
) (*acmednsproviderdto.TestProviderAccessResp, error) {
	err := acme.DNS01ProviderTestAccess(ctx, req.Kind, req.ToEntity(), req.TestDomain)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return &acmednsproviderdto.TestProviderAccessResp{}, nil
}
