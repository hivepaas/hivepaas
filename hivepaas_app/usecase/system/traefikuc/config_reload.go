package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ReloadTraefikConfig(
	ctx context.Context,
	auth *basedto.Auth,
	_ *traefikdto.ReloadTraefikConfigReq,
) (*traefikdto.ReloadTraefikConfigResp, error) {
	// Before the reload, for the reason the restart does it: the SIGHUP reaches
	// traefik's containers with nothing here to undo it.
	err := uc.recordTraefikAction(ctx, uc.db, auth, auditSectionConfigReload)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.traefikService.ReloadTraefikConfig(ctx, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &traefikdto.ReloadTraefikConfigResp{}, nil
}
