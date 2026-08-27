package appcontaineruc

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

const (
	defaultCheckPortTimeout = time.Second * 5
)

func (uc *UC) CheckAppContainerPort(
	ctx context.Context,
	auth *basedto.Auth,
	req *appcontainerdto.CheckAppContainerPortReq,
) (*appcontainerdto.CheckAppContainerPortResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	address := net.JoinHostPort(app.Key, strconv.Itoa(int(req.Port))) //nolint
	timeout := gofn.Coalesce(req.Timeout.ToDuration(), defaultCheckPortTimeout)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil && conn != nil {
		defer conn.Close()
		return &appcontainerdto.CheckAppContainerPortResp{
			Data: &appcontainerdto.CheckAppContainerPortDataResp{Open: true},
		}, nil
	}

	return &appcontainerdto.CheckAppContainerPortResp{
		Data: &appcontainerdto.CheckAppContainerPortDataResp{Open: false},
	}, nil
}
