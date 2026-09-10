package projectuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// recordProjectWrite records one write against a project.
//
// Called inside the caller's transaction and before it returns, so a record that
// cannot be written rolls the change back with it. A change that happened without
// a record is the gap the record exists to close.
//
// logType separates the three things that can happen to a project. The edits all
// share project-update, with section saying which endpoint made them, because a
// project is written from several endpoints and a type per endpoint would name
// the routing of the day it was written while filling the type filter with values
// nobody can use. Creation and removal get their own types: they are what
// somebody scanning that column is looking for, and they read as nothing at all
// folded into edits.
//
// detail may be nil, for a section with nothing to add beyond its own name.
func (uc *UC) recordProjectWrite(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	project *entity.Project,
	logType base.AuditLogType,
	source base.AuditLogSource,
	section string,
	detail *auditdetail.Builder,
) error {
	if project == nil {
		return hperrors.NewArgumentInvalidNT("audited project")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     logType,
		Scope:    base.ObjectScopeProject,
		ObjectID: project.ID,
		Source:   source,
		Auth:     auth,
		ResType:  base.ResourceTypeProject,
		ResID:    project.ID,
		ResName:  project.Name,
		Detail:   detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// photoAction names what the request did to the photo. A project has no preset
// icons, so there are only two outcomes here where an app has three.
func photoAction(req *projectdto.ProjectPhotoReq) string {
	if req.Delete {
		return "remove"
	}
	return "upload"
}
