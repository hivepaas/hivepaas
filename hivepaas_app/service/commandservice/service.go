package commandservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	BuildCommand(ctx context.Context, db database.IDB, req *BuildCommandReq) (*BuildCommandResp, error)
	GetCommand(ctx context.Context, cmdType, cmdName string) (*entity.Setting, error)
}
