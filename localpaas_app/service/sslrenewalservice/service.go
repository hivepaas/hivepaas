package sslrenewalservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	SSLRenew(ctx context.Context, db database.Tx, req *SSLRenewalReq) (*SSLRenewalResp, error)
}
