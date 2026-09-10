package projectenvuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordProjectEnvAction records one write against a project environment.
//
// Filed under the environment's own scope, so it reaches somebody reading that
// environment's log rather than only the project's.
//
// logType is a parameter rather than fixed because the two writes here are not
// the same kind of event: changing an environment's status is one more way a
// project is edited, while removing an environment takes every app inside it.
func (uc *UC) recordProjectEnvAction(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	projectEnv *entity.ProjectEnv,
	logType base.AuditLogType,
	source base.AuditLogSource,
	section string,
	detail *auditdetail.Builder,
) error {
	if projectEnv == nil {
		return hperrors.NewArgumentInvalidNT("audited project env")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     logType,
		Scope:    base.ObjectScopeProjectEnv,
		ObjectID: projectEnv.ID,
		Source:   source,
		Auth:     auth,
		ResType:  base.ResourceTypeProjectEnv,
		ResID:    projectEnv.ID,
		ResName:  projectEnv.Name,
		Detail:   detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
