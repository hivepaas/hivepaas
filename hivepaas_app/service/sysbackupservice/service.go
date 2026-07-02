package sysbackupservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	Backup(ctx context.Context, db database.Tx, req *SysBackupReq) (*SysBackupResp, error)
}
