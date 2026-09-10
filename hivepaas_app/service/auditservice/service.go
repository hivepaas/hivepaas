// Package auditservice records the actions that have to be answerable for later.
//
// It is deliberately unlike the system-error store next door. That one is
// best-effort and collapses repeats, because losing one copy of an error costs
// nothing. Here a lost entry is the whole problem: the point of the record is to
// show what happened, so Record returns an error and the caller is expected to
// abandon the action rather than perform it unrecorded.
package auditservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Entry is one action to record. The caller supplies what happened; the service
// fills in who and from where, from the auth and the request context.
type Entry struct {
	Scope    base.ObjectScopeType
	ObjectID string
	Type     base.AuditLogType
	Source   base.AuditLogSource
	Result   base.AuditLogResult

	Auth *basedto.Auth

	ResType base.ResourceType
	ResID   string
	// ResName is stored as it reads now, because the object may be renamed or
	// deleted long before anyone comes to read the entry.
	ResName string

	// Detail is free-form JSON. It must never carry a secret value.
	Detail string
}

type Service interface {
	// Record writes the entry.
	//
	// Treat a returned error as a reason to abort whatever was about to happen.
	// Performing a recordable action after its record failed leaves exactly the
	// gap the record exists to close.
	Record(ctx context.Context, db database.IDB, entry *Entry) error
}

// RecordAllowed records an action that was carried out, refusing to record one it
// cannot attribute.
//
// The refusal is the point. A usecase reaches this with the caller threaded down
// from the handler, and the way that goes wrong is not a crash but a forgotten
// assignment - somebody adds an endpoint, copies its neighbor, and the entry it
// writes names nobody. An entry naming nobody answers none of the questions the
// entry exists for, so it is better for the write to fail loudly in development
// than to succeed and leave a hole nobody sees until it matters.
//
// Only outcomes that happened belong here; a refusal is recorded with Result
// denied, by whichever gate refused it.
func RecordAllowed(ctx context.Context, svc Service, db database.IDB, entry *Entry) error {
	if entry == nil {
		return hperrors.NewArgumentInvalidNT("audit entry")
	}
	if entry.Auth == nil {
		return hperrors.NewArgumentInvalidNT("audit auth")
	}
	entry.Result = base.AuditLogResultAllowed
	if err := svc.Record(ctx, db, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
