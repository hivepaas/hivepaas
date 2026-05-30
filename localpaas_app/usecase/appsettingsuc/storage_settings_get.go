package appsettingsuc

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppStorageSettingsReq,
) (*appsettingsdto.GetAppStorageSettingsResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &appsettingsdto.StorageSettingsTransformInput{
		App:     app,
		Project: app.Project,
	}

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	input.Service = service

	// Filter out unsupported mount types
	for _, mnt := range service.Spec.TaskTemplate.ContainerSpec.Mounts {
		if gofn.Contain(supportedMountTypes, mnt.Type) {
			input.ReturningMounts = append(input.ReturningMounts, &mnt)
		}
	}

	// Load project storage settings to make sure these app settings comply with
	storageSttg, err := uc.settingRepo.GetSingle(ctx, uc.db, base.NewObjectScopeProject(app.ProjectID),
		base.SettingTypeStorageSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	input.Setting = storageSttg

	// Load reference cluster volumes as their IDs are different from their names
	if storageSttg != nil {
		storageSettings := storageSttg.MustAsStorageSettings()
		volResp, err := uc.dockerManager.VolumeListByIDs(ctx,
			storageSettings.ClusterVolumeSettings.Volumes.ToIDStringSlice())
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		for i := range volResp.Items {
			input.Volumes = append(input.Volumes, &volResp.Items[i])
		}
	}

	resp, err := appsettingsdto.TransformStorageSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppStorageSettingsResp{
		Data: resp,
	}, nil
}
