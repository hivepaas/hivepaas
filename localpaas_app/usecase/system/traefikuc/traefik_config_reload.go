package traefikuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ReloadTraefikConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.ReloadTraefikConfigReq,
) (*traefikdto.ReloadTraefikConfigResp, error) {
	err := uc.traefikService.ReloadTraefikConfig(ctx, false)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &traefikdto.ReloadTraefikConfigResp{}, nil
}
