package imagescanuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

// ExecuteImageScan runs the daily scan now.
func (uc *UC) ExecuteImageScan(
	ctx context.Context,
	_ *basedto.Auth,
	req *imagescandto.ExecuteImageScanReq,
) (*imagescandto.ExecuteImageScanResp, error) {
	req.Type = currentSettingType
	setting, err := uc.SettingRepo.GetSingle(ctx, uc.DB, req.Scope, currentSettingType, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := setting.AsImageScanSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled {
		return nil, hperrors.Wrap(hperrors.ErrImageScanOff)
	}
	job, err := uc.SettingRepo.GetSingle(ctx, uc.DB, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeImageScan),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", setting.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	task, err := uc.schedJobService.CreateSchedJobTask(job, timeNow, timeNow)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskRepo.Insert(ctx, uc.DB, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.taskQueue.ScheduleTask(ctx, task); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &imagescandto.ExecuteImageScanResp{
		Data: &imagescandto.ExecuteImageScanDataResp{Task: &basedto.ObjectIDResp{ID: task.ID}},
	}, nil
}
