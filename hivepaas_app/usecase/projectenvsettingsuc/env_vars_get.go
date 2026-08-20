package projectenvsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

func (uc *UC) GetProjectEnvEnvVars(
	ctx context.Context,
	auth *basedto.Auth,
	req *projectenvsettingsdto.GetProjectEnvEnvVarsReq,
) (*projectenvsettingsdto.GetProjectEnvEnvVarsResp, error) {
	projectEnv, err := uc.projectEnvRepo.GetByID(ctx, uc.db, req.ProjectID, req.ProjectEnvID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	settings, _, err := uc.settingRepo.List(ctx, uc.db, projectEnv.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEnvVar),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	input := &projectenvsettingsdto.EnvVarsTransformationInput{
		ProjectEnv: projectEnv,
		Vars:       settings,
	}
	resp, err := projectenvsettingsdto.TransformEnvVars(input)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &projectenvsettingsdto.GetProjectEnvEnvVarsResp{
		Data: resp,
	}, nil
}
