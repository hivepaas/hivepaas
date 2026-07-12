package hpappuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) RestartHpApp(
	ctx context.Context,
	_ *basedto.Auth,
	req *hpappdto.RestartHpAppReq,
) (*hpappdto.RestartHpAppResp, error) {
	var errCache, errDb, errMain error
	if req.RestartCacheApp {
		errCache = uc.hpAppService.RestartHpCacheSwarmService(ctx)
	}
	if req.RestartDbApp {
		errDb = uc.hpAppService.RestartHpDbSwarmService(ctx)
	}
	if req.RestartMainApp {
		errMain = uc.hpAppService.RestartHpAppSwarmService(ctx)
	}

	err := errors.Join(errMain, errDb, errCache)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &hpappdto.RestartHpAppResp{}, nil
}
