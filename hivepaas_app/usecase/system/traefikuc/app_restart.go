package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) RestartTraefik(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.RestartTraefikReq,
) (*traefikdto.RestartTraefikResp, error) {
	err := uc.traefikService.RestartTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &traefikdto.RestartTraefikResp{}, nil
}
