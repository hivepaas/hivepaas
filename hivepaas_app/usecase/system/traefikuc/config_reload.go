package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ReloadTraefikConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.ReloadTraefikConfigReq,
) (*traefikdto.ReloadTraefikConfigResp, error) {
	err := uc.traefikService.ReloadTraefikConfig(ctx, false)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &traefikdto.ReloadTraefikConfigResp{}, nil
}
