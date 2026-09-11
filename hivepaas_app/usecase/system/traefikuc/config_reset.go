package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefikuc/traefikdto"
)

func (uc *UC) ResetTraefikConfig(
	ctx context.Context,
	auth *basedto.Auth,
	_ *traefikdto.ResetTraefikConfigReq,
) (*traefikdto.ResetTraefikConfigResp, error) {
	// Before the reset, on the same terms as the other two: it acts on the files
	// traefik is watching, and there is nothing here to undo it.
	err := uc.recordTraefikAction(ctx, uc.db, auth, auditSectionConfigReset)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.traefikService.ResetTraefikConfig(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &traefikdto.ResetTraefikConfigResp{}, nil
}
