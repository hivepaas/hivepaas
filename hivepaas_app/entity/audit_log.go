package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// AuditLog is one recorded action, kept so the question "who saw this, and when"
// has an answer.
//
// Rows are append-only. Nothing updates one, and the only deletion is the
// retention sweep - see sysCleanupDBOldAuditLogs, which will not let the entries
// that matter be aged out early.
//
// Several fields duplicate data that also lives in another table on purpose. They
// are a snapshot of the moment: the user may be deleted and the setting renamed
// long before anyone reads the entry, and an entry that has become a pair of
// meaningless identifiers is no use exactly when it is needed.
type AuditLog struct {
	ID     string              `bun:",pk" json:"id"`
	Type   base.AuditLogType   `json:"type"`
	Source base.AuditLogSource `json:"source,omitempty"`
	Result base.AuditLogResult `json:"result"`

	ActorID   string           `json:"actorId"`
	ActorType base.SubjectType `json:"actorType"`
	ActorName string           `json:"actorName,omitempty"`
	// ViaAPIKey separates a person doing something from one of their API keys
	// doing it. Both carry the same user id, and they are not the same event.
	ViaAPIKey bool `json:"viaApiKey,omitempty"`
	// SessionUID identifies the session the action came from, so one compromised
	// session can be told apart from the owner's own use.
	SessionUID string `json:"sessionUid,omitempty"`

	ResType base.ResourceType `json:"resType,omitempty"`
	ResID   string            `json:"resId,omitempty"`
	ResName string            `json:"resName,omitempty"`

	// ClientIP is only as trustworthy as http_server.trusted_proxies; RemoteAddr
	// is the peer address, which a caller cannot choose. See pkg/reqinfo.
	ClientIP   string `json:"clientIp,omitempty"`
	RemoteAddr string `json:"remoteAddr,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	RequestID  string `json:"requestId,omitempty"`

	// Detail is free-form JSON for whatever a particular type needs, so a new
	// type does not need a new column. It must never hold a secret value, not
	// even shortened or hashed.
	Detail string `json:"detail,omitempty"`

	CreatedAt time.Time `bun:",default:current_timestamp" json:"createdAt"`
}

// GetID implements IDEntity interface
func (e *AuditLog) GetID() string {
	return e.ID
}
