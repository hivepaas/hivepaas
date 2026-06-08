package appdeploymentuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

func (uc *UC) GetDeploymentLogsToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentLogsTokenReq,
) (*appdeploymentdto.GetDeploymentLogsTokenResp, error) {
	token, err := uc.userService.GenerateConsoleToken(auth.User.ID, req.DeploymentID)
	if err != nil {
		return nil, apperrors.New(err).WithMsgLog("failed to generate console token")
	}

	return &appdeploymentdto.GetDeploymentLogsTokenResp{
		Data: &appdeploymentdto.DeploymentLogsTokenDataResp{
			Token: token,
		},
	}, nil
}
