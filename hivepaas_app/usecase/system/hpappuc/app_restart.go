package hpappuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) RestartHpApp(
	ctx context.Context,
	_ *basedto.Auth,
	req *hpappdto.RestartHpAppReq,
) (*hpappdto.RestartHpAppResp, error) {
	var errCache, errDb, errMain, errWorker, errAgent error
	if req.RestartCacheApp {
		errCache = uc.hpAppService.RestartHpCacheSwarmService(ctx)
	}
	if req.RestartDbApp {
		errDb = uc.hpAppService.RestartHpDbSwarmService(ctx)
	}
	if req.RestartMainApp {
		errMain = uc.hpAppService.RestartHpAppSwarmService(ctx)
	}
	if req.RestartWorkers {
		errWorker = uc.hpAppService.RestartHpWorkerSwarmService(ctx)
	}
	if req.RestartAgents {
		errAgent = uc.hpAppService.RestartHpAgentSwarmService(ctx)
	}

	err := errors.Join(errMain, errDb, errCache, errAgent, errWorker)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappdto.RestartHpAppResp{}, nil
}
