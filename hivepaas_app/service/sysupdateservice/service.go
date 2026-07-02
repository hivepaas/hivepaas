package sysupdateservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SysUpdate(ctx context.Context, db database.IDB, req *SysUpdateReq) (*SysUpdateResp, error)
}
