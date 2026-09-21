package appsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func (uc *UC) GetAppKindSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.GetAppKindSettingsReq,
) (*appsettingsdto.GetAppKindSettingsResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, false, false,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppKind, base.SettingTypeAppRouting),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", app.ID), // load app direct settings
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	kindSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind)
	secretsRevealed := false
	if kindSetting != nil && req.RevealSecrets {
		secretsRevealed, err = uc.appService.RevealSecrets(ctx, uc.db, auth, app, kindSetting)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	input := &appsettingsdto.AppKindSettingsTransformInput{
		App:            app,
		KindSetting:    kindSetting,
		RoutingSetting: settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting),
		MaskSecrets:    !secretsRevealed,
	}

	resp, err := appsettingsdto.TransformAppKindSettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &appsettingsdto.GetAppKindSettingsResp{
		Data: resp,
	}, nil
}
