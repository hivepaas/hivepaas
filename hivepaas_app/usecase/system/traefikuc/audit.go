package traefikuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// The sections traefik's three actions are recorded under.
//
// Prefixed, because they share AuditLogTypeHivePaaSAction with what the HivePaaS
// app does, and that already uses "restart" and "config-reload" for its own. The
// two are not the same event - one cycles the dashboard, the other cycles every
// app's ingress - and section is what a reader has to tell them apart by.
const (
	auditSectionRestart      = "traefik-restart"
	auditSectionConfigReload = "traefik-config-reload"
	auditSectionConfigReset  = "traefik-config-reset"
)

// recordTraefikAction records something done to the running traefik.
//
// Filed under the hivepaas scope, like the traefik settings and the HivePaaS
// actions next door: traefik is a real app, but none of this is read as an app's
// history - it is the install's ingress being cycled, and the question the entry
// answers is what was done to this install.
//
// No resource is named, the way hpappuc names none. These act on a running
// service rather than on a row, there is no id to point at, and the listing only
// renders a resource that has one. The section says which service and which act.
//
// No detail either, because none of the three takes anything: the section is the
// whole of what there is to say. The HivePaaS restart next door carries one only
// because it is asked which services to cycle.
func (uc *UC) recordTraefikAction(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	section string,
) error {
	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeHivePaaSAction,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIAction,
		Section: section,
		Auth:    auth,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
