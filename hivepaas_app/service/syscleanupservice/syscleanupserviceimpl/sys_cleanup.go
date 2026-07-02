package syscleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/syscleanupservice"
)

type sysCleanupData struct {
	*syscleanupservice.SysCleanupReq
	TaskOutput *entity.TaskSystemCleanupOutput
}

func (s *service) Cleanup(
	ctx context.Context,
	db database.Tx,
	req *syscleanupservice.SysCleanupReq,
) (resp *syscleanupservice.SysCleanupResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	data := &sysCleanupData{
		SysCleanupReq: req,
		TaskOutput: &entity.TaskSystemCleanupOutput{
			DBCleanup:      &entity.DBCleanupOutput{},
			ClusterCleanup: &entity.ClusterCleanupOutput{},
			BackupCleanup:  &entity.BackupCleanupOutput{},
			CacheCleanup:   &entity.CacheCleanupOutput{},
			FileCleanup:    &entity.FileCleanupOutput{},
		},
	}
	if data.LogStore == nil {
		data.LogStore = tasklog.NewLocalStore(fmt.Sprintf("task:%v:log", req.Task.ID))
	}
	resp = &syscleanupservice.SysCleanupResp{
		TaskOutput: data.TaskOutput,
	}

	var errs []error

	// Cleanup DB objects
	errs = append(errs, s.sysCleanupDB(ctx, db, data))

	// Cleanup unused cluster data (docker)
	errs = append(errs, s.sysCleanupCluster(ctx, data))

	// Cleanup old backup files
	errs = append(errs, s.sysCleanupBackups(ctx, db, data))

	// Cleanup outdated cache files
	errs = append(errs, s.sysCleanupCache(ctx, db, data))

	// Cleanup orphaned files
	errs = append(errs, s.sysCleanupFiles(ctx, data))

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	return resp, errors.Join(errs...)
}
