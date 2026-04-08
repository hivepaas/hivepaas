package traefikuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) RestartTraefik(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.RestartTraefikReq,
) (*traefikdto.RestartTraefikResp, error) {
	err := uc.traefikService.RestartTraefikSwarmService(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &traefikdto.RestartTraefikResp{}, nil
}
