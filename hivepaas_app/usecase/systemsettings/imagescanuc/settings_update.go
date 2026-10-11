package imagescanuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/imagescanuc/imagescandto"
)

const (
	currentSettingType   = base.SettingTypeImageScan
	imageScanSettingName = "Image scanning settings"
	imageScanJobName     = "Image scan job"
	// A daily scan that failed - the server fetching its DB - is run again
	// once, a few minutes later.
	imageScanJobMaxRetry   = 1
	imageScanJobRetryDelay = timeutil.Duration(5 * time.Minute)
)

// UpdateImageScanSettings stores the configuration, makes the cluster match it
// and schedules the daily scan by it.
func (uc *UC) UpdateImageScanSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagescandto.UpdateImageScanSettingsReq,
) (*imagescandto.UpdateImageScanSettingsResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	updateData := &updateSettingData{NewSettings: req.ToEntity()}
	var applied *imagescanservice.SettingApplyResp

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: imageScanSettingName,
		Load: func(ctx context.Context, db database.Tx, data *settings.UpdateUniqueSettingData) error {
			updateData.UpdateUniqueSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return hperrors.Wrap(pData.Setting.SetData(updateData.NewSettings))
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			if err := uc.saveJob(ctx, db, updateData); err != nil {
				return err
			}
			// Make the cluster match what was just saved. Apply is idempotent, so
			// a failure here leaves a stored configuration the next save retries.
			applyResp, applyErr := uc.imageScanService.Apply(ctx, db, &imagescanservice.SettingApplyReq{
				Setting:       pData.Setting,
				TriggerUserID: auth.UserID(),
				ScanResources: req.ScanResources(),
			})
			applied = applyResp
			return hperrors.Wrap(applyErr)
		},
	})
	if err != nil {
		// The records went with the transaction; what provisioning made in docker
		// did not, and nothing else would take it down.
		if applied != nil && applied.Cleanup != nil {
			err = errors.Join(err, applied.Cleanup(context.WithoutCancel(ctx)))
		}
		return nil, hperrors.Wrap(err)
	}

	// A task can be picked up only once its row exists, which is once the
	// transaction has committed.
	if applied != nil && len(applied.Tasks) > 0 {
		if err = uc.taskQueue.ScheduleTask(ctx, applied.Tasks...); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &imagescandto.UpdateImageScanSettingsResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewSettings        *entity.ImageScanSettings
	JobSetting         *entity.Setting
	JobScheduleChanges bool
}

// loadSettingData reads what is stored, makes it when it is not, carries over
// what the client does not own, and loads the daily scan's job.
func (uc *UC) loadSettingData(
	ctx context.Context,
	db database.Tx,
	req *imagescandto.UpdateImageScanSettingsReq,
	data *updateSettingData,
) error {
	setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, currentSettingType, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	timeNow := timeutil.NowUTC()
	if setting == nil {
		setting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType,
			Type:      currentSettingType,
			Status:    base.SettingStatusActive,
			Name:      imageScanSettingName,
			Version:   entity.CurrentImageScanSettingsVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = setting

	current, err := setting.AsImageScanSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	// What the addon wrote is the server's, not the client's.
	data.NewSettings.ServerAppID, data.NewSettings.RunnerAppID = current.ServerAppID, current.RunnerAppID
	data.JobScheduleChanges = !current.Schedule.Equal(&data.NewSettings.Schedule)

	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeImageScan),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", setting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if jobSetting == nil {
		jobSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType,
			Type:      base.SettingTypeSchedJob,
			Kind:      string(base.SchedJobTypeImageScan),
			Status:    base.SettingStatusDisabled,
			Name:      imageScanJobName,
			Version:   entity.CurrentSchedJobVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		jobSetting.MustSetData(&entity.SchedJob{
			JobType:       base.SchedJobTypeImageScan,
			Schedule:      &entity.SchedJobSchedule{},
			TargetSetting: entity.ObjectID{ID: setting.ID},
			MaxRetry:      imageScanJobMaxRetry,
			RetryDelay:    imageScanJobRetryDelay,
		})
		data.JobScheduleChanges = true
	}
	data.JobSetting = jobSetting
	return nil
}

// saveJob writes the daily scan's job: on while scanning is, with the
// settings' schedule and notification, scheduled again when either changed.
func (uc *UC) saveJob(ctx context.Context, db database.Tx, data *updateSettingData) error {
	job := data.JobSetting
	status := gofn.If(data.NewSettings.Enabled, base.SettingStatusActive, base.SettingStatusDisabled)
	statusChanges := job.Status != status
	job.Status = status
	schedJob := job.MustAsSchedJob()
	schedJob.Schedule = &data.NewSettings.Schedule
	schedJob.Notification = data.NewSettings.Notification
	job.MustSetData(schedJob)
	job.UpdateVer++
	job.UpdatedAt = timeutil.NowUTC()
	if err := uc.SettingRepo.Upsert(ctx, db, job, entity.SettingUpsertingConflictCols,
		entity.SettingUpsertingUpdateCols); err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, job,
		data.JobScheduleChanges || statusChanges))
}
