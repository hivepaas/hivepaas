package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ResetTraefikConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.ResetTraefikConfigReq,
) (*traefikdto.ResetTraefikConfigResp, error) {
	err := uc.traefikService.ResetTraefikConfig(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &traefikdto.ResetTraefikConfigResp{}, nil
}
