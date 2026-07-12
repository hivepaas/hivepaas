package schedjobuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
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
			if err := uc.isSchedJobFeatureEnabledInApp(ctx, db, data.ScopeApp); err != nil {
				return apperrors.Wrap(err)
			}
			if err := uc.checkPermissionPipeToApp(ctx, db, auth, schedJob); err != nil {
				return apperrors.Wrap(err)
			}
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
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	// If command output is piped to another app, need to check permission
	hasPerm, err := uc.PermissionManager.CheckAccess(ctx, db, auth, &permission.AccessCheck{
		ResourceModule:     base.ResourceModuleProject,
		ResourceType:       base.ResourceTypeApp,
		ResourceID:         targetApp.ID,
		ParentResourceType: base.ResourceTypeProject,
		ParentResourceID:   targetApp.ProjectID,
		Action:             base.ActionTypeWrite,
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
