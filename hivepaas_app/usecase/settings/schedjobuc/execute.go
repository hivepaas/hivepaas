package schedjobuc

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{
		SkipLoadingRefObjects: true,
		AfterLoading: func(
			ctx context.Context,
			db database.IDB,
			data *settings.GetSettingData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, data.ScopeApp); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	task, err := uc.schedJobService.CreateSchedJobTask(resp.Data, time.Time{}, timeutil.NowUTC())
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskRepo.Insert(ctx, uc.DB, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTask(ctx, task)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.ExecuteSchedJobResp{
		Data: &schedjobdto.ExecuteSchedJobDataResp{
			Task: &basedto.ObjectIDResp{ID: task.ID},
		},
	}, nil
}
