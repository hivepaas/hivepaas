package auditlogdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type ListAuditLogTypeReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewListAuditLogTypeReq() *ListAuditLogTypeReq {
	return &ListAuditLogTypeReq{}
}

type ListAuditLogTypeResp struct {
	Meta *basedto.ListMeta   `json:"meta"`
	Data []base.AuditLogType `json:"data"`
}
