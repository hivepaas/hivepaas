package projectsettingsuc

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

// recordProjectUpdate records one write against a project.
//
// Called inside the caller's transaction and before it returns, so a record that
// cannot be written rolls the change back with it. A change that happened without
// a record is the gap the record exists to close.
//
// section says which endpoint's slice of the project was written - its details,
// its user accesses, its env vars. It is a detail field rather than an audit type
// of its own because a project is written from several endpoints, and a type per
// endpoint would name the routing of the day it was written while filling the
// type filter with values nobody can use.
//
// detail may be nil, for a section with nothing to add beyond its own name.
func (uc *UC) recordProjectUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	project *entity.Project,
	source base.AuditLogSource,
	section string,
	detail *auditdetail.Builder,
) error {
	if project == nil {
		return hperrors.NewArgumentInvalid("audited project")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeProjectUpdate,
		Scope:    base.ObjectScopeProject,
		ObjectID: project.ID,
		Source:   source,
		Section:  section,
		Auth:     auth,
		ResType:  base.ResourceTypeProject,
		ResID:    project.ID,
		ResName:  project.Name,
		Detail:   detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
