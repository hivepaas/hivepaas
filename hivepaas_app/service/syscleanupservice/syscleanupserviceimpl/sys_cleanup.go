package syscleanupserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
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
	defer safego.RecoverTo(&err)

	data := &sysCleanupData{
		SysCleanupReq: req,
		TaskOutput: &entity.TaskSystemCleanupOutput{
			DBCleanup:      &entity.DBCleanupOutput{},
			ClusterCleanup: &entity.ClusterCleanupOutput{},
			CacheCleanup:   &entity.CacheCleanupOutput{},
			FileCleanup:    &entity.FileCleanupOutput{},
		},
	}
	if req.Scope == nil {
		req.Scope = entity.NewObjectScopeGlobal()
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

	// Bring the system apps to their settings, and check OBI on the nodes:
	// before the cluster's cleanup, which would otherwise prune the image of
	// one whose service is gone, only for it to be pulled again.
	errs = append(errs, s.sysSyncSystemApps(ctx, db, data))

	// Cleanup unused cluster data (docker)
	errs = append(errs, s.sysCleanupCluster(ctx, data))

	// Cleanup outdated cache files
	errs = append(errs, s.sysCleanupCache(ctx, db, data))

	// Cleanup orphaned files
	errs = append(errs, s.sysCleanupFiles(ctx, data))

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	return resp, errors.Join(errs...)
}
