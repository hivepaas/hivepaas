package appsettingsuc

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/mount"
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

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Load project storage settings to make sure these app settings comply with
	storageSttg, err := uc.settingRepo.GetSingle(ctx, uc.db, base.NewSettingScopeProject(app.ProjectID),
		base.SettingTypeStorageSettings, true)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.Wrap(err)
	}
	var storageSettings *entity.StorageSettings
	if storageSttg != nil {
		storageSettings = storageSttg.MustAsStorageSettings()
	} else {
		storageSettings = &entity.StorageSettings{}
	}

	// Filter out unsupported mount types
	returningMounts := make([]mount.Mount, 0, len(service.Spec.TaskTemplate.ContainerSpec.Mounts))
	for _, mnt := range service.Spec.TaskTemplate.ContainerSpec.Mounts {
		if gofn.Contain(supportedMountTypes, mnt.Type) {
			returningMounts = append(returningMounts, mnt)
		}
	}
	service.Spec.TaskTemplate.ContainerSpec.Mounts = returningMounts

	resp, err := appsettingsdto.TransformStorageSettings(app, storageSettings, service)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppStorageSettingsResp{
		Data: resp,
	}, nil
}
