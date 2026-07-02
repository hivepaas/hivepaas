package appsettingsuc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

const (
	defaultCheckPortTimeout = time.Second * 5
)

func (uc *UC) CheckAppContainerPort(
	ctx context.Context,
	auth *basedto.Auth,
	req *appsettingsdto.CheckAppContainerPortReq,
) (*appsettingsdto.CheckAppContainerPortResp, error) {
	app, err := uc.appRepo.GetByID(ctx, uc.db, req.ProjectID, req.AppID,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return nil, apperrors.New(err)
	}

	address := net.JoinHostPort(app.Key, strconv.Itoa(int(req.Port))) //nolint
	timeout := gofn.Coalesce(req.Timeout.ToDuration(), defaultCheckPortTimeout)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil && conn != nil {
		defer conn.Close()
		return &appsettingsdto.CheckAppContainerPortResp{
			Data: &appsettingsdto.CheckAppContainerPortDataResp{Open: true},
		}, nil
	}

	return &appsettingsdto.CheckAppContainerPortResp{
		Data: &appsettingsdto.CheckAppContainerPortDataResp{Open: false},
	}, nil
}
