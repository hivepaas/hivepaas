package schedjobuc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

const (
	scriptLenThreshold = 2048
)

func (uc *UC) CreateSchedJob(
	ctx context.Context,
	auth *basedto.Auth,
	req *schedjobdto.CreateSchedJobReq,
) (*schedjobdto.CreateSchedJobResp, error) {
	req.Type = currentSettingType
	schedJob := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Name,
		VerifyingRefIDs: schedJob.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, req.Scope.App); err != nil {
				return apperrors.Wrap(err)
			}
			if err := uc.checkPermissionPipeToApp(ctx, db, auth, schedJob); err != nil {
				return apperrors.Wrap(err)
			}
			// Calculate upserting script objects if the scripts are too long
			upsertingScripts := uc.calcUpsertingScriptSettings(pData.Setting, schedJob, nil)
			pData.UpsertingSettings = append(pData.UpsertingSettings, upsertingScripts...)

			pData.Setting.Kind = string(schedJob.JobType)
			if err := pData.Setting.SetData(schedJob); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		AfterPersisting: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			err := uc.taskQueue.ScheduleTasksForSchedJob(ctx, db, pData.Setting, false)
			if err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &schedjobdto.CreateSchedJobResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) checkPermissionPipeToApp(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	schedJob *entity.SchedJob,
) error {
	cmdOutput := schedJob.CommandOutput
	if cmdOutput == nil || cmdOutput.PipeToApp == nil {
		return nil
	}

	targetAppID := cmdOutput.PipeToApp.TargetApp.ID
	targetApp, err := uc.AppService.LoadApp(ctx, db, "", targetAppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If command output is piped to another app, need to check permission
	hasPerm, err := uc.PermissionManager.CheckAccess(ctx, db, auth, &permission.AppAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		AppID:           targetApp.ID,
		ParentID:        targetApp.ParentID,
		ProjectID:       targetApp.ProjectID,
		ProjectEnv:      targetApp.ProjectEnvID,
	})
	if err != nil {
		return apperrors.Wrap(err)
	}
	if !hasPerm {
		return apperrors.Wrap(apperrors.ErrUnauthorized)
	}
	return nil
}

func (uc *UC) isSchedJobFeatureEnabledInApp(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) error {
	if app == nil {
		return nil
	}
	featureSetting, err := uc.SettingRepo.GetSingle(ctx, db, app.GetObjectScope(),
		base.SettingTypeAppFeatures, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.Wrap(err)
	}
	var featureSettings *entity.AppFeatureSettings
	if featureSetting != nil {
		featureSettings = featureSetting.MustAsAppFeatureSettings()
	} else {
		featureSettings = &entity.AppFeatureSettings{}
		entity.InitAppFeatureSettingsDefault(featureSettings)
	}
	if featureSettings.SchedJobSettings != nil && !featureSettings.SchedJobSettings.Enabled {
		return apperrors.Wrap(apperrors.ErrFeatureDisabled).WithParam("Name", "scheduled-job")
	}
	return nil
}

func (uc *UC) calcUpsertingScriptSettings(
	jobSetting *entity.Setting,
	newSchedJob *entity.SchedJob,
	oldSchedJob *entity.SchedJob,
) (scripts []*entity.Setting) {
	timeNow := time.Now()
	var oldScriptID, oldPipeToAppScriptID string

	if oldSchedJob != nil {
		if oldSchedJob.Command != nil {
			oldScriptID = oldSchedJob.Command.Script.ID
		}
		if oldSchedJob.CommandOutput != nil && oldSchedJob.CommandOutput.PipeToApp != nil {
			oldPipeToAppScriptID = oldSchedJob.CommandOutput.PipeToApp.Command.Script.ID
		}
	}

	if newSchedJob != nil {
		cmd := newSchedJob.Command
		if cmd != nil && len(cmd.Script.Value) > scriptLenThreshold {
			scriptSetting := newScriptSetting(jobSetting, timeNow)
			if oldScriptID != "" {
				scriptSetting.ID = oldScriptID
				oldScriptID = ""
			}
			scriptSetting.MustSetData(&entity.Script{Data: cmd.Script.Value})
			cmd.Script.ID = scriptSetting.ID
			cmd.Script.Value = ""
			scripts = append(scripts, scriptSetting)
		}

		cmdOut := newSchedJob.CommandOutput
		if cmdOut != nil && cmdOut.PipeToApp != nil && len(cmdOut.PipeToApp.Command.Script.Value) > scriptLenThreshold {
			scriptSetting := newScriptSetting(jobSetting, timeNow)
			if oldPipeToAppScriptID != "" {
				scriptSetting.ID = oldPipeToAppScriptID
				oldPipeToAppScriptID = ""
			}
			scriptSetting.MustSetData(&entity.Script{Data: cmdOut.PipeToApp.Command.Script.Value})
			cmdOut.PipeToApp.Command.Script.ID = scriptSetting.ID
			cmdOut.PipeToApp.Command.Script.Value = ""
			scripts = append(scripts, scriptSetting)
		}
	}

	if oldScriptID != "" { // the ID is unused, delete the linked script
		scriptSetting := newScriptSetting(jobSetting, timeNow)
		scriptSetting.ID = oldScriptID
		scriptSetting.DeletedAt = timeNow
		scripts = append(scripts, scriptSetting)
	}
	if oldPipeToAppScriptID != "" { // the ID is unused, delete the linked script
		scriptSetting := newScriptSetting(jobSetting, timeNow)
		scriptSetting.ID = oldPipeToAppScriptID
		scriptSetting.DeletedAt = timeNow
		scripts = append(scripts, scriptSetting)
	}

	return scripts
}

func newScriptSetting(jobSetting *entity.Setting, timeNow time.Time) *entity.Setting {
	return &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     jobSetting.Scope,
		ObjectID:  jobSetting.ObjectID,
		Type:      base.SettingTypeScript,
		Status:    base.SettingStatusActive,
		Name:      "script of job: " + jobSetting.Name,
		Version:   entity.CurrentScriptVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
}
