package auditloguc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
)

func (uc *UC) ListAuditLog(
	ctx context.Context,
	auth *basedto.Auth,
	req *auditlogdto.ListAuditLogReq,
) (*auditlogdto.ListAuditLogResp, error) {
	if req.Scope == nil {
		req.Scope = entity.NewObjectScopeGlobal()
	}
	targetScope := req.Scope
	switch {
	case req.ProjectID != "":
		targetScope = entity.NewObjectScopeProject(req.ProjectID)
	case req.ProjectEnvID != "":
		projectID, env := projecthelper.ParseProjectEnvID(req.ProjectEnvID)
		if projectID != "" && env != "" {
			targetScope = entity.NewObjectScopeProjectEnv(projectID, env)
		}
	case req.AppID != "":
		targetScope = entity.NewObjectScopeApp(req.AppID, "", "", "")
	}
	targetScope.NoInherited = req.Scope.NoInherited
	if req.ScopeOnly {
		targetScope.NoInherited = true
	}

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
	if !req.FromDate.IsZero() {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.created_at >= ?",
			req.FromDate.ToTime()))
	}
	if !req.ToDate.IsZero() {
		listOpts = append(listOpts, bunex.SelectWhereIn("audit_log.created_at < ?",
			req.ToDate.AddDate(0, 0, 1).ToTime()))
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

	auditLogs, pagingMeta, err := uc.auditLogRepo.List(ctx, uc.db, targetScope, &req.Paging, listOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	refObjects.AddObjectScope(targetScope)
	err = uc.loadAuditLogRefData(ctx, uc.db, auditLogs, &refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := auditlogdto.TransformAuditLogs(auditLogs, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &auditlogdto.ListAuditLogResp{
		Meta: &basedto.ListMeta{Page: pagingMeta},
		Data: resp,
	}, nil
}

func (uc *UC) loadAuditLogRefData(
	ctx context.Context,
	db database.IDB,
	auditLogs []*entity.AuditLog,
	refObjects **entity.RefObjects,
) error {
	if len(auditLogs) == 0 {
		return nil
	}
	refIDs := &entity.RefObjectIDs{}
	for _, auditLog := range auditLogs {
		if auditLog == nil {
			continue
		}
		refIDs.AddRefIDs(auditLog.GetRefObjectIDs())
	}
	err := uc.settingService.LoadRefObjectsByIDsSkipMissing(ctx, db, refObjects, nil, false, refIDs)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
