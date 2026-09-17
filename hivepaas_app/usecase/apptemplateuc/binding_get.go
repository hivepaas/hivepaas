package apptemplateuc

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

func (uc *UC) GetAppTemplateBinding(
	ctx context.Context,
	_ *basedto.Auth,
	req *apptemplatedto.GetAppTemplateBindingReq,
) (*apptemplatedto.GetAppTemplateBindingResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app.ProjectEnvID != req.ProjectEnvID {
		return nil, hperrors.Wrap(hperrors.ErrAppNotFound).WithParam("Name", req.AppID)
	}

	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, app.GetObjectScope(), base.SettingTypeAppTemplate, true,
		bunex.SelectWhere("setting.object_id = ?", app.ID),
	)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.NewNotFound("App template")
		}
		return nil, hperrors.Wrap(err)
	}
	data, err := setting.AsAppTemplateSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &apptemplatedto.GetAppTemplateBindingResp{
		Data: apptemplatedto.TransformAppTemplateBinding(data),
	}, nil
}
