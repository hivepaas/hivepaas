package appdeploymentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordAppAction records something being done to an app, named by id.
//
// This package reaches a deployment without reaching the app that owns it - it
// holds deployment repositories and nothing else - so the entry carries the app
// id without its name. That is a real loss: ResName exists because an app can be
// renamed or deleted long before anyone reads the entry. Loading the app for the
// sake of the name would put an extra query on the cancel path, so what stands in
// for it is the deployment id in the detail, which identifies the thing acted on
// even when the app around it has moved on.
func (uc *UC) recordAppAction(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	appID string,
	section string,
	detail *auditdetail.Builder,
) error {
	if appID == "" {
		return hperrors.NewArgumentInvalidNT("audited app")
	}
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:     base.AuditLogTypeAppUpdate,
		Scope:    base.ObjectScopeApp,
		ObjectID: appID,
		Source:   base.AuditLogSourceAPIAction,
		Auth:     auth,
		ResType:  base.ResourceTypeApp,
		ResID:    appID,
		Detail:   detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
