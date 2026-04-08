package appuc

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppStorageSettingsReq,
) (*appdto.UpdateAppStorageSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppStorageSettingsData{}
		err := uc.loadAppStorageSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppStorageSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppStorageSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppStorageSettingsResp{}, nil
}

type updateAppStorageSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppStorageSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppStorageSettings(
	req *appdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) {
	uc.prepareUpdatingAppStorageMounts(req, data)
}

func (uc *UC) prepareUpdatingAppStorageMounts(
	req *appdto.UpdateAppStorageSettingsReq,
	data *updateAppStorageSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	if len(req.Mounts) == 0 {
		containerSpec.Mounts = nil
		return
	}

	currMounts := make(map[string]*mount.Mount, len(containerSpec.Mounts))
	for i := range containerSpec.Mounts {
		mnt := &containerSpec.Mounts[i]
		// Use type and source to identify a mount
		currMounts[fmt.Sprintf("type:%v:src:%v", mnt.Type, mnt.Source)] = mnt
	}

	containerSpec.Mounts = make([]mount.Mount, 0, len(req.Mounts))
	for _, reqMount := range req.Mounts {
		mnt := currMounts[fmt.Sprintf("type:%v:src:%v", reqMount.Type, reqMount.Source)]
		if mnt == nil {
			mnt = &mount.Mount{
				Type:   reqMount.Type,
				Source: reqMount.Source,
			}
		}
		mnt.Target = reqMount.Target
		mnt.Consistency = reqMount.Consistency
		containerSpec.Mounts = append(containerSpec.Mounts, *mnt)
	}
}

func (uc *UC) applyAppStorageSettings(
	ctx context.Context,
	data *updateAppStorageSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
