package appsettingsuc

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppStorageSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppStorageSettingsReq,
) (*appsettingsdto.GetAppStorageSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	service, err := uc.clusterService.ServiceInspect(ctx, app.ServiceID, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &appsettingsdto.StorageSettingsTransformInput{
		App:                app,
		Service:            service,
		MountKeyCalculator: uc.calcMountKey,
	}

	resp, err := appsettingsdto.TransformStorageSettings(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppStorageSettingsResp{
		Data: resp,
	}, nil
}

func (uc *UC) calcMountKey(mnt *mount.Mount) string {
	key := fmt.Sprintf("type:%v:src:%v:target:%v", mnt.Type, mnt.Source, mnt.Target)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}
