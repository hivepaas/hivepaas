package auditlogdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

type ListAuditLogTypeReq struct {
}

func NewListAuditLogTypeReq() *ListAuditLogTypeReq {
	return &ListAuditLogTypeReq{}
}

type ListAuditLogTypeResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []base.AuditLogType `json:"data"`
}
