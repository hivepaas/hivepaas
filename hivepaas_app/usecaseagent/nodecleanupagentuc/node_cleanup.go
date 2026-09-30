package nodecleanupagentuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodecleanupagentuc/nodecleanupagentdto"
)

func (uc *UC) NodeCleanup(
	ctx context.Context,
	req *nodecleanupagentdto.NodeCleanupReq,
) (*nodecleanupagentdto.NodeCleanupResp, error) {
	if req.CleanupSettings == nil || !req.CleanupSettings.Enabled {
		return &nodecleanupagentdto.NodeCleanupResp{}, nil
	}

	if req.TaskExecData == nil {
		req.TaskExecData = &queue.TaskExecData{
			Task: &entity.Task{ID: req.TaskID},
		}
	} else if req.Task == nil {
		req.Task = &entity.Task{ID: req.TaskID}
	}

	// What a build killed half way left in this node's temporary directory.
	_, _ = fileutil.RemoveDatedTempDirs(base.BaseTempDirDefault,
		time.Now().AddDate(0, 0, -fileutil.TempDirRetentionDays))

	resp, err := uc.clusterCleanupService.Cleanup(ctx, &req.ClusterCleanupReq)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	out := resp.Output
	if out == nil {
		out = &entity.ClusterNodeCleanupOutput{}
	}

	return &nodecleanupagentdto.NodeCleanupResp{
		ClusterNodeCleanupOutput: *out,
	}, nil
}
