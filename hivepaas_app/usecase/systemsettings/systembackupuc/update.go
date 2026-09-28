package systembackupuc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

const (
	currentSettingType  = base.SettingTypeSystemBackup
	backupSettingName   = "System backup settings"
	backupJobName       = "System backup job"
	backupJobMaxRetry   = 1
	backupJobRetryDelay = timeutil.Duration(time.Second * 60)
)

func (uc *UC) UpdateSystemBackup(
	ctx context.Context,
	auth *basedto.Auth,
	req *systembackupdto.UpdateSystemBackupReq,
) (*systembackupdto.UpdateSystemBackupResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	updateData := &updateSettingData{
		NewBackup: req.ToEntity(),
	}
	persistingData := &persistingSettingData{}

	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: backupSettingName,
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

	return &systembackupdto.UpdateSystemBackupResp{}, nil
}

type updateSettingData struct {
	*settings.UpdateUniqueSettingData
	NewBackup          *entity.SystemBackup
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
	req *systembackupdto.UpdateSystemBackupReq,
	data *updateSettingData,
) error {
	backupSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSystemBackup, false,
		bunex.SelectFor("UPDATE OF setting"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if backupSetting == nil {
		timeNow := timeutil.NowUTC()
		backupSetting = &entity.Setting{
			ID:        gofn.Must(ulid.NewStringULID()),
			Scope:     req.Scope.ScopeType,
			Type:      base.SettingTypeSystemBackup,
			Status:    base.SettingStatusActive,
			Name:      backupSettingName,
			Version:   entity.CurrentSystemBackupVersion,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		}
	}
	data.Setting = backupSetting

	backup, err := backupSetting.AsSystemBackup()
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.JobScheduleChanges = !backup.Schedule.Equal(&data.NewBackup.Schedule)

	req.KeepMaskedSecrets(data.NewBackup, backup)
	if err = uc.checkNewBackup(ctx, db, req.Auth, data.NewBackup); err != nil {
		return hperrors.Wrap(err)
	}

	// Load sched job of the backup
	jobSetting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.data->'targetSetting'->>'id' = ?", backupSetting.ID),
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
			Status:      base.SettingStatusActive,
			Name:        backupJobName,
			Inheritable: true,
			Version:     entity.CurrentSchedJobVersion,
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
		}
		schedJob := &entity.SchedJob{
			JobType:       base.SchedJobTypeSystemBackup,
			Schedule:      &entity.SchedJobSchedule{},
			TargetSetting: entity.ObjectID{ID: backupSetting.ID},
			MaxRetry:      backupJobMaxRetry,
			RetryDelay:    backupJobRetryDelay,
		}
		jobSetting.MustSetData(schedJob)
	}
	data.JobSetting = jobSetting

	return nil
}

// checkNewBackup refuses a repository that is not a global one, and a spec that
// holds secrets from a person without the capability to reveal them.
//
// The gate is a mount's, the capability alone: the backup puts the secrets in a
// repository, as a mount puts one in a container, and hands the caller nothing
// in the clear - so the operator's flag on what the API returns does not apply.
func (uc *UC) checkNewBackup(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	backup *entity.SystemBackup,
) error {
	if backup.TargetRepository.ID != "" {
		repo, err := uc.SettingRepo.GetByID(ctx, db, nil, base.SettingTypeBackupRepo, backup.TargetRepository.ID, false)
		if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
			return hperrors.Wrap(err)
		}
		if err = checkTargetRepository(repo); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if !backup.IncludeSpec || !specmodel.SecretsMode(backup.SpecSecrets).RevealsSecrets() {
		return nil
	}
	err := uc.PermissionManager.AuthorizeSecretMount(ctx, db, auth, &permission.RevealSubject{
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIAction,
		ResType: base.ResourceTypeSetting,
		ResName: fmt.Sprintf("system backup spec (%s)", backup.SpecSecrets),
	})
	return hperrors.Wrap(err)
}

// checkTargetRepository is a backup repository at the global scope.
func checkTargetRepository(repo *entity.Setting) error {
	if repo == nil {
		return hperrors.NewNotFound("Backup repository")
	}
	if repo.Scope != base.ObjectScopeGlobal {
		return hperrors.NewArgumentInvalid("targetRepository").
			WithExtraDetail("the system backup goes into a repository of the global scope")
	}
	return nil
}

func (uc *UC) preparePersistingData(
	req *systembackupdto.UpdateSystemBackupReq,
	updateData *updateSettingData,
	persistingData *persistingSettingData,
) error {
	// Set new backup settings
	persistingData.Setting.Status = req.Status
	err := persistingData.Setting.SetData(updateData.NewBackup)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// Update backup job
	jobSetting := updateData.JobSetting
	jobSetting.Status = gofn.If(persistingData.Setting.Status == base.SettingStatusActive,
		base.SettingStatusActive, base.SettingStatusDisabled)
	jobSetting.Kind = string(base.SchedJobTypeSystemBackup)
	persistingData.JobSetting = jobSetting

	backupJob := jobSetting.MustAsSchedJob()
	backupJob.Schedule = &updateData.NewBackup.Schedule
	backupJob.Notification = updateData.NewBackup.Notification
	jobSetting.MustSetData(backupJob)

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
