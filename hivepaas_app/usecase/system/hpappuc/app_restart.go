package hpappuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappuc/hpappdto"
)

func (uc *UC) RestartHpApp(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappdto.RestartHpAppReq,
) (*hpappdto.RestartHpAppResp, error) {
	// Before the restarts, because they go straight to swarm and there is no
	// transaction to roll back: writing the entry first is the only order that
	// cannot leave a restart of the install unrecorded. Which services were asked
	// for goes in - restarting the main app is a different event from cycling the
	// database underneath it.
	err := uc.recordHpAppAction(ctx, uc.db, auth, "restart", auditdetail.New().
		Set("mainApp", req.RestartMainApp).
		Set("workers", req.RestartWorkers).
		Set("agents", req.RestartAgents).
		Set("dbApp", req.RestartDbApp).
		Set("cacheApp", req.RestartCacheApp))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

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

	err = errors.Join(errMain, errDb, errCache, errAgent, errWorker)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappdto.RestartHpAppResp{}, nil
}
