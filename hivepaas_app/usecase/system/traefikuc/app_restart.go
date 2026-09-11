package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) RestartTraefik(
	ctx context.Context,
	auth *basedto.Auth,
	_ *traefikdto.RestartTraefikReq,
) (*traefikdto.RestartTraefikResp, error) {
	// Before the restart, because it goes straight to swarm and there is no
	// transaction to roll back: writing the entry first is the only order that
	// cannot leave every app's ingress being cycled unrecorded. It is also the
	// order that survives the restart cutting this request short.
	err := uc.recordTraefikAction(ctx, uc.db, auth, auditSectionRestart)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.traefikService.RestartTraefikSwarmService(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &traefikdto.RestartTraefikResp{}, nil
}
