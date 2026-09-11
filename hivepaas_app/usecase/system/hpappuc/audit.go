package hpappuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// recordHpAppAction records something done to the running HivePaaS install.
//
// Filed under the hivepaas scope, the same one its settings use: there is no
// project or app to put an upgrade or a restart under, and an entry filed under
// one would be invisible to somebody reading the install's own log.
//
// section says which action it was - version-update, restart, config-reload.
//
// No resource is named, on purpose. These act on the whole install rather than on
// a row, there is no id to point at, and the listing only renders a resource that
// has one - so inventing a resource type here would be a value that never shows
// and never answers anything. The hivepaas scope already says what was acted on.
func (uc *UC) recordHpAppAction(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	section string,
	detail *auditdetail.Builder,
) error {
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeHivePaaSAction,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIAction,
		Section: section,
		Auth:    auth,
		Detail:  detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
