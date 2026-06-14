package builduc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
	"github.com/localpaas/localpaas/localpaas_app/tasks/queue"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/builduc/builddto"
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
			return apperrors.Wrap(err)
		}
		cachesDeleted = resp.TaskOutput.ClusterCleanup.BuildCachesDeleted
		spaceReclaimed = resp.TaskOutput.ClusterCleanup.SpaceReclaimed
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
