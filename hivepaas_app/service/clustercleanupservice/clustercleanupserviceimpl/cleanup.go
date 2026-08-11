package clustercleanupserviceimpl

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clustercleanupservice"
)

type clusterCleanupData struct {
	*clustercleanupservice.ClusterCleanupReq
	Output *entity.ClusterNodeCleanupOutput
}

func (s *service) Cleanup(
	ctx context.Context,
	req *clustercleanupservice.ClusterCleanupReq,
) (resp *clustercleanupservice.ClusterCleanupResp, err error) {
	defer funcutil.EnsureNoPanic(&err)

	data := &clusterCleanupData{
		ClusterCleanupReq: req,
		Output:            &entity.ClusterNodeCleanupOutput{},
	}
	if data.LogStore == nil {
		data.LogStore = tasklog.NewLocalStore(fmt.Sprintf("task:%v:log", req.Task.ID))
	}
	resp = &clustercleanupservice.ClusterCleanupResp{
		Output: data.Output,
	}

	err = s.cleanupCluster(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, nil
}
