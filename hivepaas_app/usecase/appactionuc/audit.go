package appactionuc

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

// recordAppAction records something being done to a running app - a deploy, a
// restart, a stop, a file pushed into a container.
//
// Written before the thing happens, which is the opposite of how the settings
// flows do it and is deliberate. Those run inside a transaction, so a record that
// fails takes the change down with it. There is no transaction here: Docker
// cannot be rolled back, so the only order that guarantees a record is to write
// it first and abandon the action if it fails. The cost is that an action which
// then fails still leaves an entry - and for an action that is the better way to
// be wrong, because "who asked for this" is the question, and an attempt that
// errored is worth as much as one that worked.
//
// section says which action it was. It is a detail field rather than an audit
// type of its own so every write against an app stays under one type - see
// AuditLogTypeAppUpdate.
func (uc *UC) recordAppAction(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	app *entity.App,
	source base.AuditLogSource,
	section string,
	detail *auditdetail.Builder,
) error {
	if app == nil {
		return hperrors.NewArgumentInvalid("audited app")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: app.ID,
		Source:   source,
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
