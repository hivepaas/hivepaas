package auditloguc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

func (uc *UC) ListAuditLogType(
	ctx context.Context,
	auth *basedto.Auth,
	req *auditlogdto.ListAuditLogTypeReq,
) (*auditlogdto.ListAuditLogTypeResp, error) {
	return &auditlogdto.ListAuditLogTypeResp{
		Data: base.AllAuditLogTypes,
	}, nil
}
