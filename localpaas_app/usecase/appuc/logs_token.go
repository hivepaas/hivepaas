package appuc

import (
	"context"
	"fmt"
	"time"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/entity/cacheentity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

const (
	consoleUIDByteLen = 12
	consoleTokenExp   = 30 * time.Second
)

func (uc *UC) GetAppLogsToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.GetAppLogsTokenReq,
) (*appdto.GetAppLogsTokenResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.New(apperrors.ErrUnavailable).
			WithMsgLog("service not exist for app")
	}

	token := fmt.Sprintf("app:%s:log:%s", app.ID, gofn.RandTokenAsHex(consoleUIDByteLen))

	err = uc.consoleTicketRepo.Set(ctx, token, &cacheentity.ConsoleTicket{
		AppID:    req.AppID,
		TargetID: app.ServiceID,
	}, consoleTokenExp)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.GetAppLogsTokenResp{
		Data: &appdto.AppLogsTokenDataResp{
			Token: token,
		},
	}, nil
}
