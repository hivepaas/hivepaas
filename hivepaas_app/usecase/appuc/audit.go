package appuc

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

// recordAppUpdate records one write against an app.
//
// Called inside the caller's transaction and before it returns, so a record that
// cannot be written rolls the change back with it. A change that happened without
// a record is the gap the record exists to close.
//
// section says which settings tab was written - routing, resources, deployment,
// and the rest. It is a detail field rather than an audit type of its own because
// an app is written from a dozen endpoints, and a type per endpoint would name
// the routing of the day it was written while filling the type filter with values
// nobody can use.
//
// detail may be nil, for a section with nothing to add beyond its own name. Some
// sections carry the list of fields that moved and some do not, and that is not
// an oversight: the ones backed by a stored settings row have a before and an
// after to compare, while the ones that write straight to the Swarm service -
// resources, container, networks - have no stored row to diff.
func (uc *UC) recordAppUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	section string,
	detail *auditdetail.Builder,
) error {
	if app == nil {
		return hperrors.NewArgumentInvalidNT("audited app")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   base.AuditLogSourceAPIUpdate,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    app.ID,
		ResName:  app.Name,
		Detail:   detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
