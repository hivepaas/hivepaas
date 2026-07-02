package appcopyservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	CopyApp(ctx context.Context, db database.Tx, req *AppCopyReq) (*AppCopyResp, error)
}
