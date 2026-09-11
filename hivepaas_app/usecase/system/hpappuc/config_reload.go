package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) ReloadHpAppConfig(
	ctx context.Context,
	auth *basedto.Auth,
	_ *hpappdto.ReloadHpAppConfigReq,
) (*hpappdto.ReloadHpAppConfigResp, error) {
	// Before the reload, for the reason the restart above does it: it fans out to
	// the other replicas with nothing here to undo it.
	err := uc.recordHpAppAction(ctx, uc.db, auth, "config-reload", nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.hpAppService.ReloadHpAppConfig(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappdto.ReloadHpAppConfigResp{}, nil
}
