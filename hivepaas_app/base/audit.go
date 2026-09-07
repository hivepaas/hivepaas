package base

// AuditLogType is what happened, named after the action rather than after the
// endpoint it arrived on.
//
// An endpoint-shaped name (say "get-reveal") ties recorded history to the routing
// of the day it was written: move the action to another route later and the same
// event starts being recorded under a different name, so any question asked across
// that boundary quietly gets a wrong answer. Where the action came in is recorded
// separately, as AuditLogSource.
type AuditLogType string

const (
	// AuditLogTypeSecretReveal records a stored secret being handed out in the
	// clear, whether or not the caller was allowed to have it.
	AuditLogTypeSecretReveal AuditLogType = "secret-reveal"

	// AuditLogTypeAPIKeyCreate records a long-lived credential being minted.
	AuditLogTypeAPIKeyCreate AuditLogType = "api-key-create"

	// AuditLogTypeAPIKeyRevoke records a key being taken out of use. Keys that
	// disappear without a record are the same problem as secrets handed out
	// without one, so the cascade that follows a revoked capability writes these.
	AuditLogTypeAPIKeyRevoke AuditLogType = "api-key-revoke" //nolint:gosec // G101: an event name

	// AuditLogTypeSecuritySettingsUpdate records a change to the operator-level
	// security switches. One of them decides whether stored secrets may leave the
	// server at all, so the change is worth as much as the reveals it permits:
	// without this, a switch flipped on and back off leaves the reveals in between
	// looking like they were always allowed.
	AuditLogTypeSecuritySettingsUpdate AuditLogType = "security-settings-update"

	// AuditLogTypeRoutingChangeConfirm records somebody vouching that a routing
	// change left HivePaaS reachable, which is what stops it being undone.
	AuditLogTypeRoutingChangeConfirm AuditLogType = "routing-change-confirm"

	// AuditLogTypeRoutingChangeRevert records a routing change being undone on
	// request, before its deadline. The automatic undo is not recorded here: it
	// has no caller to attribute, and its record is the task row that ran it.
	AuditLogTypeRoutingChangeRevert AuditLogType = "routing-change-revert"
)

// AuditLogSource is the way in - which endpoint, or which subsystem.
type AuditLogSource string

const (
	// AuditLogSourceAPIGet is a plain GET asked to include the secrets.
	AuditLogSourceAPIGet AuditLogSource = "api-get"

	// AuditLogSourceAPICreate is a create endpoint.
	AuditLogSourceAPICreate AuditLogSource = "api-create"

	// AuditLogSourceCapabilityRevoked is the system acting on its own, following
	// through on a capability somebody took away.
	AuditLogSourceCapabilityRevoked AuditLogSource = "capability-revoked"

	// AuditLogSourceAPIUpdate is an update endpoint.
	AuditLogSourceAPIUpdate AuditLogSource = "api-update"
)

// AuditLogResult says whether the action was permitted.
//
// The refusals are the more interesting half: a denied attempt is the only signal
// that somebody is trying doors, and it is invisible if only successes are kept.
type AuditLogResult string

const (
	AuditLogResultAllowed AuditLogResult = "allowed"
	AuditLogResultDenied  AuditLogResult = "denied"
)
