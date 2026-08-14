package builduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/builduc/builddto"
)

func (uc *UC) ClearBuildCache(
	ctx context.Context,
	auth *basedto.Auth,
	req *builddto.ClearBuildCacheReq,
) (*builddto.ClearBuildCacheResp, error) {
	cleanupReq := &syscleanupservice.SysCleanupReq{
		TaskExecData: &queue.TaskExecData{
			Task: &entity.Task{
				ID: "fake-task-id",
			},
		},
		Scope: entity.NewObjectScopeGlobal(),
		SysCleanupSettings: &entity.SystemCleanup{
			ClusterCleanup: entity.SystemClusterCleanup{
				Enabled:         true,
				PruneBuildCache: true,
			},
		},
		CleanupClusterBuildCache: base.CleanupFlagForce,
	}

	cachesDeleted := 0
	spaceReclaimed := uint64(0)
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		resp, err := uc.sysCleanupService.Cleanup(ctx, db, cleanupReq)
		if err != nil {
			return apperrors.Wrap(err)
		}
		if resp.TaskOutput.ClusterCleanup == nil {
			resp.TaskOutput.ClusterCleanup = &entity.ClusterCleanupOutput{}
		}
		for _, nodeReport := range resp.TaskOutput.ClusterCleanup.Nodes {
			cachesDeleted += nodeReport.BuildCachesDeleted
			spaceReclaimed += nodeReport.SpaceReclaimed
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &builddto.ClearBuildCacheResp{
		Data: &builddto.ClearBuildCacheDataResp{
			CachesDeleted:  cachesDeleted,
			SpaceReclaimed: spaceReclaimed,
		},
	}, nil
}
