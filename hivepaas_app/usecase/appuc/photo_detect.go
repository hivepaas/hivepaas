package appuc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/apphelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

func (uc *UC) DetectAppPhoto(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.DetectAppPhotoReq,
) (*appdto.DetectAppPhotoResp, error) {
	// NOTE: no need to load project/env relations as passing `requireActive = false`
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var imageName string
	if app.ServiceID != "" {
		swarmSvc, err := uc.dockerManager.ServiceInspect(ctx, app.ServiceID)
		if err == nil && swarmSvc != nil && swarmSvc.Service.Spec.TaskTemplate.ContainerSpec != nil {
			imageName = swarmSvc.Service.Spec.TaskTemplate.ContainerSpec.Image
		}
	}

	var photoURL string
	iconName := apphelper.DetectAppIcon(app.Name, imageName)
	if iconName != "" {
		photoURL = filepath.Join(config.Current.HttpPathStaticIcons(), fmt.Sprintf("%s.svg", iconName))
	}

	return &appdto.DetectAppPhotoResp{
		Data: &appdto.DetectAppPhotoDataResp{URL: photoURL},
	}, nil
}
