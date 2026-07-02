package builduc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
			Task: &entity.Task{},
		},
		SysCleanupSettings: &entity.SystemCleanup{
			ClusterCleanup: entity.SystemClusterCleanup{
				Enabled: true,
			},
		},
		CleanupClusterBuildCache: syscleanupservice.CleanupFlagForce,
	}

	cachesDeleted := 0
	spaceReclaimed := uint64(0)
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		resp, err := uc.sysCleanupService.Cleanup(ctx, db, cleanupReq)
		if err != nil {
			return apperrors.New(err)
		}
		cachesDeleted = resp.TaskOutput.ClusterCleanup.BuildCachesDeleted
		spaceReclaimed = resp.TaskOutput.ClusterCleanup.SpaceReclaimed
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &builddto.ClearBuildCacheResp{
		Data: &builddto.ClearBuildCacheDataResp{
			CachesDeleted:  cachesDeleted,
			SpaceReclaimed: spaceReclaimed,
		},
	}, nil
}
