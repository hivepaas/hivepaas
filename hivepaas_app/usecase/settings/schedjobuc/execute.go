package schedjobuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

func (uc *UC) ExecuteSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.ExecuteSchedJobReq,
) (*schedjobdto.ExecuteSchedJobResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{
		SkipLoadingRefObjects: true,
		AfterLoading: func(
			ctx context.Context,
			db database.IDB,
			data *settings.GetSettingData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	task, err := uc.schedJobService.CreateSchedJobTask(resp.Data, timeNow, timeNow)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &schedjobdto.ExecuteSchedJobResp{
		Data: &schedjobdto.ExecuteSchedJobDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}
