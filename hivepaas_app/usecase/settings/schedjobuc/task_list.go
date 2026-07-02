package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) ListSchedJobTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ListSchedJobTaskReq,
) (*schedjobdto.ListSchedJobTaskResp, error) {
	req.Type = currentSettingType
	jobSetting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.JobID, false)
	if err != nil {
		return nil, apperrors.New(err)
	}

	listResp, err := uc.taskService.ListTask(ctx, uc.DB, &taskservice.ListTaskReq{
		TargetIDs: []string{jobSetting.ID},
		Statuses:  req.Status,
		Search:    req.Search,
		Paging:    req.Paging,
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	resp, err := taskdto.TransformTasks(listResp.Tasks, listResp.TaskInfoMap)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &schedjobdto.ListSchedJobTaskResp{
		Meta: &basedto.ListMeta{Page: listResp.PagingMeta},
		Data: resp,
	}, nil
}
