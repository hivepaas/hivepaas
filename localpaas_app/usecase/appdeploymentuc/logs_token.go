package appdeploymentuc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

const (
	consoleUIDByteLen = 12
	consoleTokenExp   = 30 * time.Second
)

func (uc *UC) GetDeploymentLogsToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdeploymentdto.GetDeploymentLogsTokenReq,
) (*appdeploymentdto.GetDeploymentLogsTokenResp, error) {
	deployment, err := uc.deploymentRepo.GetByID(ctx, uc.db, req.AppID, req.DeploymentID)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	token := fmt.Sprintf("app-%s-log-%s", deployment.AppID, gofn.RandTokenAsHex(consoleUIDByteLen))
	key := strings.ReplaceAll(token, "-", ":")

	err = uc.consoleTicketRepo.Set(ctx, key, &cacheentity.ConsoleTicket{
		AppID:    req.AppID,
		TargetID: req.DeploymentID,
	}, consoleTokenExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdeploymentdto.GetDeploymentLogsTokenResp{
		Data: &appdeploymentdto.DeploymentLogsTokenDataResp{
			Token: token,
		},
	}, nil
}
