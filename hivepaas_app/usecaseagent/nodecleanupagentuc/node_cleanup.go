package nodecleanupagentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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

	resp, err := uc.clusterCleanupService.Cleanup(ctx, &req.ClusterCleanupReq)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	out := resp.Output
	if out == nil {
		out = &entity.ClusterNodeCleanupOutput{}
	}

	return &nodecleanupagentdto.NodeCleanupResp{
		ClusterNodeCleanupOutput: *out,
	}, nil
}
