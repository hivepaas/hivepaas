package sysupdateservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	SysUpdate(ctx context.Context, db database.IDB, req *SysUpdateReq) (*SysUpdateResp, error)
}
