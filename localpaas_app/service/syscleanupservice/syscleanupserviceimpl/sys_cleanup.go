package syscleanupserviceimpl

import (
	"context"
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/funcutil"
	"github.com/localpaas/localpaas/localpaas_app/service/syscleanupservice"
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

	resp = &syscleanupservice.SysCleanupResp{}
	data := &sysCleanupData{
		SysCleanupReq: req,
		TaskOutput: &entity.TaskSystemCleanupOutput{
			DBCleanup:      &entity.DBCleanupOutput{},
			ClusterCleanup: &entity.ClusterCleanupOutput{},
			FileCleanup:    &entity.FileCleanupOutput{},
		},
	}

	// Cleanup DB objects
	err1 := s.sysCleanupDB(ctx, db, data)

	// Cleanup unused cluster data (docker)
	err2 := s.sysCleanupCluster(ctx, data)

	// Cleanup orphaned files
	err3 := s.sysCleanupFiles(ctx, data)

	// Assign back the result output
	data.Task.MustSetOutput(data.TaskOutput)

	return resp, errors.Join(err1, err2, err3)
}
