package sslrenewalservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	SSLRenew(ctx context.Context, db database.Tx, req *SSLRenewalReq) (*SSLRenewalResp, error)
}
