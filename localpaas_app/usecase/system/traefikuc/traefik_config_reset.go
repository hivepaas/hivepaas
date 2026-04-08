package traefikuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ResetTraefikConfig(
	ctx context.Context,
	_ *basedto.Auth,
	_ *traefikdto.ResetTraefikConfigReq,
) (*traefikdto.ResetTraefikConfigResp, error) {
	err := uc.traefikService.ResetTraefikConfig(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &traefikdto.ResetTraefikConfigResp{}, nil
}
