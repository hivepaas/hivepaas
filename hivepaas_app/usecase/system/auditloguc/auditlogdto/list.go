package auditlogdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListAuditLogReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`

	Type       []base.AuditLogType   `json:"-" mapstructure:"type"`
	Source     []base.AuditLogSource `json:"-" mapstructure:"source"`
	Result     []base.AuditLogResult `json:"-" mapstructure:"result"`
	ActorID    []string              `json:"-" mapstructure:"actorId"`
	ResourceID []string              `json:"-" mapstructure:"resourceId"`
	Search     string                `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListAuditLogReq() *ListAuditLogReq {
	return &ListAuditLogReq{
		Paging: basedto.Paging{
			// Default paging if unset by client
			Sort: basedto.Orders{{Direction: basedto.DirectionDesc, ColumnName: "created_at"}},
		},
	}
}

func (req *ListAuditLogReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAuditLogResp struct {
	Meta *basedto.ListMeta `json:"meta"`
	Data []*AuditLogResp   `json:"data"`
}

func TransformAuditLogs(
	auditLogs []*entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp []*AuditLogResp, err error) {
	resp = make([]*AuditLogResp, 0, len(auditLogs))
	for _, auditLog := range auditLogs {
		item, err := TransformAuditLog(auditLog, refObjects)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, item)
	}
	return resp, nil
}
