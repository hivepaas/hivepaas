// Package clusteraudit records the operations carried out on the cluster.
//
// It is one function shared by the four cluster usecases rather than a helper
// copied into each. They are separate packages with separate UC types and nothing
// else in common, and four copies of the same entry-building code is four places
// for the scope or the type to drift apart - which, for a record whose job is to
// be searchable, is the failure that matters.
package clusteraudit

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// Target is what the operation acted on.
//
// ID and Name are empty for an operation that has no single subject - purging
// the build cache across every node, reconciling a whole table. ResType still
// says which part of the cluster it was.
type Target struct {
	ResType base.ResourceType
	ID      string
	Name    string
}

// Record writes one cluster operation.
//
// The scope is global because that is what a cluster is: there is no project or
// app to file it under, and an entry filed under one would be invisible to
// somebody reading the cluster's own log.
//
// section names the operation. detail may be nil.
func Record(
	ctx context.Context,
	svc auditservice.Service,
	db database.IDB,
	auth *basedto.Auth,
	target Target,
	section string,
	detail *auditdetail.Builder,
) error {
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, svc, db, &auditservice.Entry{
		Type:    base.AuditLogTypeClusterUpdate,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIAction,
		Auth:    auth,
		ResType: target.ResType,
		ResID:   target.ID,
		ResName: target.Name,
		Detail:  detail.Set("section", section).String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
