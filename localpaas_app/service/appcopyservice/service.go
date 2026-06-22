package appcopyservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	CopyApp(ctx context.Context, db database.Tx, req *AppCopyReq) (*AppCopyResp, error)
}
