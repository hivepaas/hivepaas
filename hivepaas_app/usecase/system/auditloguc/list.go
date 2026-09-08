package auditloguc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/auditloguc/auditlogdto"
)

func (uc *UC) ListAuditLog(
	ctx context.Context,
	auth *basedto.Auth,
	req *auditlogdto.ListAuditLogReq,
) (*auditlogdto.ListAuditLogResp, error) {
	var listOpts []bunex.SelectQueryOption
	if len(req.Type) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.type IN (?)", req.Type...))
	}
	if len(req.Source) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.source IN (?)", req.Source...))
	}
	if len(req.Result) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.result IN (?)", req.Result...))
	}
	if len(req.ActorID) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.actor_id IN (?)", req.ActorID...))
	}
	if len(req.ResourceID) > 0 {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.res_id IN (?)", req.ResourceID...))
	}
	if req.Search != "" {
		keyword := bunex.MakeLikeOpStr(req.Search, true)
		listOpts = append(listOpts,
			bunex.SelectWhereGroup(
				bunex.SelectWhere("audit_log.actor_name ILIKE ?", keyword),
				bunex.SelectWhereOr("audit_log.res_name ILIKE ?", keyword),
				bunex.SelectWhereOr("audit_log.client_ip ILIKE ?", keyword),
				bunex.SelectWhereOr("audit_log.remote_addr ILIKE ?", keyword),
			),
		)
	}

	auditLogs, pagingMeta, err := uc.auditLogRepo.List(ctx, uc.db, req.Scope, &req.Paging, listOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// TODO: implement this when need ref objects
	refObjects := entity.NewRefObjects()

	resp, err := auditlogdto.TransformAuditLogs(auditLogs, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &auditlogdto.ListAuditLogResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}
