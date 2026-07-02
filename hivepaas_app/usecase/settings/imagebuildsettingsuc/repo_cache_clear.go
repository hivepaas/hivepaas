package imagebuildsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) ClearRepoCache(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.ClearRepoCacheReq,
) (*imagebuildsettingsdto.ClearRepoCacheResp, error) {
	cleanupReq := &syscleanupservice.SysCleanupReq{
		TaskExecData: &queue.TaskExecData{
			Task: &entity.Task{},
		},
		SysCleanupSettings: &entity.SystemCleanup{
			CacheCleanup: entity.SystemCacheCleanup{
				Enabled: true,
			},
		},
		CleanupCacheRepo: syscleanupservice.CleanupFlagForce,
	}

	filesDeleted := 0
	spaceReclaimed := uint64(0)
	err := transaction.Execute(ctx, uc.DB, func(db database.Tx) error {
		resp, err := uc.sysCleanupService.Cleanup(ctx, db, cleanupReq)
		if err != nil {
			return apperrors.New(err)
		}
		filesDeleted = resp.TaskOutput.CacheCleanup.RepoCacheFilesDeleted
		spaceReclaimed = resp.TaskOutput.CacheCleanup.RepoCacheSpaceReclaimed
		return nil
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &imagebuildsettingsdto.ClearRepoCacheResp{
		Data: &imagebuildsettingsdto.ClearRepoCacheDataResp{
			FilesDeleted:   filesDeleted,
			SpaceReclaimed: spaceReclaimed,
		},
	}, nil
}
