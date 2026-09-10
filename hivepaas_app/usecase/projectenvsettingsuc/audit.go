package projectenvsettingsuc

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

// recordProjectEnvUpdate records one write against a project environment.
//
// Filed under the project-update type and the environment's own scope: it is a
// change to a project, and the scope is what puts it in front of somebody reading
// that environment's log rather than the project's.
//
// Called inside the caller's transaction, so a record that cannot be written
// rolls the change back with it.
func (uc *UC) recordProjectEnvUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	projectEnv *entity.ProjectEnv,
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
		Type:     base.AuditLogTypeProjectUpdate,
		Scope:    base.ObjectScopeProjectEnv,
		ObjectID: projectEnv.ID,
		Source:   base.AuditLogSourceAPIUpdate,
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
