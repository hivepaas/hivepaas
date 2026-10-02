package registryauthrenewaluc

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
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/registryauthrenewaluc/registryauthrenewaldto"
)

const (
	currentSettingType   = base.SettingTypeRegistryAuthRenewal
	renewalSettingName   = "Registry auth renewal settings"
	renewalJobName       = "Registry auth renewal job"
	renewalJobMaxRetry   = 1
	renewalJobRetryDelay = timeutil.Duration(time.Second * 60)
)

func (uc *UC) UpdateRegistryAuthRenewal(
	ctx context.Context,
	auth *basedto.Auth,
	req *registryauthrenewaldto.UpdateRegistryAuthRenewalReq,
) (*registryauthrenewaldto.UpdateRegistryAuthRenewalResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	updateData := &updateSettingData{
		NewRenewal: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: renewalSettingName,
		Load: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
		) error {
			updateData.UpdateUniqueSettingData = data
			return uc.loadSettingData(ctx, db, req, updateData)
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			persistingData.PersistingSettingData = pData
			return uc.preparePersistingData(req, updateData, persistingData)
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return uc.postPersisting(ctx, db, updateData, persistingData)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &registryauthrenewaldto.UpdateRegistryAuthRenewalResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewRenewal         *entity.RegistryAuthRenewal
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
	req *registryauthrenewaldto.UpdateRegistryAuthRenewalReq,
	data *updateSettingData,
) error {
	renewalSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeRegistryAuthRenewal, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if renewalSetting == nil {
		timeNow := timeutil.NowUTC()
		renewalSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType,
			Type:      base.SettingTypeRegistryAuthRenewal,
			Status:    base.SettingStatusActive,
			Name:      renewalSettingName,
			Version:   entity.CurrentRegistryAuthRenewalVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = renewalSetting

	renewal, err := renewalSetting.AsRegistryAuthRenewal()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.JobScheduleChanges = !renewal.Schedule.Equal(&data.NewRenewal.Schedule)

	// Load sched job of the renewal
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeRegistryAuthRenewal),
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", renewalSetting.ID),
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if jobSetting == nil {
		timeNow := timeutil.NowUTC()
		jobSetting = &entity.Setting{
			ID:          gofn.Must(ulid.NewStringULID()),
			Scope:       req.Scope.ScopeType,
			Type:        base.SettingTypeSchedJob,
			Kind:        string(base.SchedJobTypeRegistryAuthRenewal),
			Status:      base.SettingStatusActive,
			Name:        renewalJobName,
			Inheritable: true,
			Version:     entity.CurrentSchedJobVersion,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		schedJob := &entity.SchedJob{
			JobType:       base.SchedJobTypeRegistryAuthRenewal,
			Schedule:      &entity.SchedJobSchedule{},
			TargetSetting: entity.ObjectID{ID: renewalSetting.ID},
			MaxRetry:      renewalJobMaxRetry,
			RetryDelay:    renewalJobRetryDelay,
		}
		jobSetting.MustSetData(schedJob)
	}
	data.JobSetting = jobSetting

	return nil
}

func (uc *UC) preparePersistingData(
	req *registryauthrenewaldto.UpdateRegistryAuthRenewalReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new renewal settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewRenewal)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update renewal job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.SchedJobTypeRegistryAuthRenewal)
	persistingData.JobSetting = jobSetting

	renewalJob := jobSetting.MustAsSchedJob()
	renewalJob.Schedule = &updateData.NewRenewal.Schedule
	renewalJob.Notification = updateData.NewRenewal.Notification
	jobSetting.MustSetData(renewalJob)

	return nil
}

func (uc *UC) postPersisting(
	ctx context.Context,
	db database.Tx,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Persist the sched job updates
	err := uc.SettingRepo.Update(ctx, db, persistingData.JobSetting)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, updateData.JobSetting, updateData.JobScheduleChanges)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
