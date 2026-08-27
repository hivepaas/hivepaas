package periodicjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/taskservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/periodicjobuc/periodicjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/taskuc/taskdto"
)

func (uc *UC) ListPeriodicJobTask(
	ctx context.Context,
	auth *basedto.Auth,
	req *periodicjobdto.ListPeriodicJobTaskReq,
) (*periodicjobdto.ListPeriodicJobTaskResp, error) {
	req.Type = currentSettingType
	jobSetting, err := uc.GetSettingByID(ctx, uc.DB, &req.BaseSettingReq, req.JobID, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	listResp, err := uc.taskService.ListTask(ctx, uc.DB, &taskservice.ListTaskReq{
		TargetIDs: []string{jobSetting.ID},
		Statuses:  req.Status,
		Search:    req.Search,
		Paging:    req.Paging,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := taskdto.TransformTasks(listResp.Tasks, listResp.TaskInfoMap)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &periodicjobdto.ListPeriodicJobTaskResp{
		Meta: &basedto.ListMeta{Page: listResp.PagingMeta},
		Data: resp,
	}, nil
}
