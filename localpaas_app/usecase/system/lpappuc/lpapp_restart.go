package lpappuc

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/system/lpappuc/lpappdto"
)

func (uc *UC) RestartLpApp(
	ctx context.Context,
	_ *basedto.Auth,
	req *lpappdto.RestartLpAppReq,
) (*lpappdto.RestartLpAppResp, error) {
	var errCache, errDb, errMain error
	if req.RestartCacheApp {
		errCache = uc.lpAppService.RestartLpCacheSwarmService(ctx)
	}
	if req.RestartDbApp {
		errDb = uc.lpAppService.RestartLpDbSwarmService(ctx)
	}
	if req.RestartMainApp {
		errMain = uc.lpAppService.RestartLpAppSwarmService(ctx)
	}

	err := errors.Join(errMain, errDb, errCache)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &lpappdto.RestartLpAppResp{}, nil
}
