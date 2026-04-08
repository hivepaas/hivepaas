package sslrenewaluc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/sslrenewaluc/sslrenewaldto"
)

const (
	currentSettingType   = base.SettingTypeSSLRenewal
	renewalSettingName   = "SSL renewal settings"
	renewalJobName       = "SSL renewal job"
	renewalJobMaxRetry   = 1
	renewalJobRetryDelay = timeutil.Duration(time.Second * 60)
)

func (uc *UC) UpdateSSLRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslrenewaldto.UpdateSSLRenewalReq,
) (*sslrenewaldto.UpdateSSLRenewalResp, error) {
	req.Type = currentSettingType
	updateData := &updateSettingData{
		NewRenewal: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName: renewalSettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
		) error {
			updateData.UpdateSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			persistingData.PersistingSettingData = pData
			return uc.preparePersistingData(req, updateData, persistingData)
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return uc.postPersisting(ctx, db, updateData, persistingData)
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslrenewaldto.UpdateSSLRenewalResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateSettingData
	NewRenewal         *entity.SSLRenewal
	JobSetting         *entity.Setting
	JobScheduleChanges bool
}

type persistingSettingData struct {
	*settings.PersistingSettingData
	JobSetting *entity.Setting
}

func (uc *UC) loadSettingData(
	ctx context.Context,
	db database.Tx,
	req *sslrenewaldto.UpdateSSLRenewalReq,
	data *updateSettingData,
) error {
	renewalSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSSLRenewal, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Setting = renewalSetting

	renewal, err := renewalSetting.AsSSLRenewal()
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.JobScheduleChanges = renewal.ScheduleInterval != data.NewRenewal.ScheduleInterval ||
		renewal.ScheduleFrom != data.NewRenewal.ScheduleFrom

	// Load cron job of the cleanup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeCronJob, false,
		bunex.SelectWhere("setting.kind = ?", base.CronJobTypeSSLRenewal),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", renewalSetting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	if jobSetting == nil {
		timeNow := timeutil.NowUTC()
		jobSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType(),
			Type:      base.SettingTypeCronJob,
			Kind:      string(base.CronJobTypeSSLRenewal),
			Status:    base.SettingStatusActive,
			Name:      renewalJobName,
			Version:   entity.CurrentCronJobVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
		cronJob := &entity.CronJob{
			CronType:      base.CronJobTypeSSLRenewal,
			Schedule:      &entity.CronJobSchedule{},
			TargetSetting: entity.ObjectID{ID: renewalSetting.ID},
			MaxRetry:      renewalJobMaxRetry,
			RetryDelay:    renewalJobRetryDelay,
		}
		jobSetting.MustSetData(cronJob)
	}
	data.JobSetting = jobSetting

	return nil
}

func (uc *UC) preparePersistingData(
	req *sslrenewaldto.UpdateSSLRenewalReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new cleanup settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewRenewal)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// Update renewal job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.CronJobTypeSSLRenewal)
	persistingData.JobSetting = jobSetting

	renewalJob := jobSetting.MustAsCronJob()
	renewalJob.Schedule.Interval = updateData.NewRenewal.ScheduleInterval
	renewalJob.Schedule.InitialTime = updateData.NewRenewal.ScheduleFrom
	renewalJob.Schedule.OnChange(updateData.JobScheduleChanges) // call this to handle if the schedule changes
	jobSetting.MustSetData(renewalJob)

	return nil
}

func (uc *UC) postPersisting(
	ctx context.Context,
	db database.Tx,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Persist the cron job updates
	err := uc.SettingRepo.Update(ctx, db, persistingData.JobSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTasksForCronJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
