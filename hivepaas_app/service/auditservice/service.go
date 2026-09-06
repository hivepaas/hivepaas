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
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Entry is one action to record. The caller supplies what happened; the service
// fills in who and from where, from the auth and the request context.
type Entry struct {
	Type   base.AuditLogType
	Source base.AuditLogSource
	Result base.AuditLogResult

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
