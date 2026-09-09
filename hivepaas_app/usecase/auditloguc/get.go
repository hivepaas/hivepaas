package auditloguc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

func (uc *UC) GetAuditLog(
	ctx context.Context,
	auth *basedto.Auth,
	req *auditlogdto.GetAuditLogReq,
) (*auditlogdto.GetAuditLogResp, error) {
	// Scoped, like the listing. An id is not an authorisation: these entries carry
	// secret reveals, refused attempts and client addresses, and reading one from
	// outside its scope would make the boundary the listing enforces decorative.
	auditLog, err := uc.auditLogRepo.GetByID(ctx, uc.db, req.Scope, req.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	refObjects.AddObjectScope(req.Scope)
	err = uc.loadAuditLogRefData(ctx, uc.db, []*entity.AuditLog{auditLog}, &refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := auditlogdto.TransformAuditLogDetail(auditLog, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &auditlogdto.GetAuditLogResp{
		Data: resp,
	}, nil
}
