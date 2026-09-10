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

	// AuditLogTypeSettingCreate, AuditLogTypeSettingUpdate,
	// AuditLogTypeSettingStatusUpdate and AuditLogTypeSettingDelete record the
	// lifecycle of a stored setting - a registry credential, an SSH key, a backup
	// repository, and the twenty-odd others.
	//
	// Four types for all of them, not four per kind. Which kind it was is already
	// on the entry as ResType, and folding it into the name instead would put
	// eighty-odd values in the type filter, which is the same as having no filter.
	//
	// Creation is recorded alongside the rest because a trail that shows a setting
	// changed and then deleted, with no record of it ever appearing, reads as if
	// something is missing from the log.
	AuditLogTypeSettingCreate AuditLogType = "setting-create"
	AuditLogTypeSettingUpdate AuditLogType = "setting-update"

	// AuditLogTypeSettingStatusUpdate is separate from a plain update because it
	// covers the switches that decide whether a setting is in force at all -
	// status, expiry, inheritable, default - and "who disabled this" is a question
	// asked on its own.
	AuditLogTypeSettingStatusUpdate AuditLogType = "setting-status-update"

	AuditLogTypeSettingDelete AuditLogType = "setting-delete"

	// AuditLogTypeProjectUpdate and AuditLogTypeAppUpdate record a change to a
	// project or an app itself, as opposed to a setting stored under it.
	//
	// One type each, covering every write. A project is changed from several
	// endpoints - its own details, its user accesses, its env vars - and an app
	// from a dozen more, one per settings tab. A type per endpoint would name the
	// routing of the day it was written and put twenty values in a filter nobody
	// can then use; which part was written is in the entry's detail, under
	// "section", where it can be read without being filtered on.
	AuditLogTypeProjectUpdate AuditLogType = "project-update"
	AuditLogTypeAppUpdate     AuditLogType = "app-update"

	// AuditLogTypeAppCreate and AuditLogTypeAppDelete stand apart from
	// app-update, which covers the edits, because an app appearing and an app
	// being taken away are what somebody scanning the type column is looking for.
	// Folded into app-update they would be indistinguishable from a change of
	// replica count, and a trail that shows an app edited and then never seen
	// again reads as if something is missing from it.
	AuditLogTypeAppCreate AuditLogType = "app-create"
	AuditLogTypeAppDelete AuditLogType = "app-delete"

	// AuditLogTypeProjectCreate and AuditLogTypeProjectDelete are the same split,
	// one level up. Removing a project takes its environments and every app in
	// them, which makes it the largest single act the API offers and the one most
	// worth being able to find by scanning one column.
	AuditLogTypeProjectCreate AuditLogType = "project-create"
	AuditLogTypeProjectDelete AuditLogType = "project-delete"

	// AuditLogTypeProjectEnvDelete records an environment being removed, which
	// takes every app inside it.
	//
	// It is not folded into project-update the way the environment's other writes
	// are. "What was destroyed" is a question asked on its own and answered by
	// scanning the type column, and an entry that says project-update cannot tell
	// a rename from an environment and its apps being taken away.
	AuditLogTypeProjectEnvDelete AuditLogType = "project-env-delete"
)

var AllAuditLogTypes = []AuditLogType{
	AuditLogTypeSecretReveal,
	AuditLogTypeAPIKeyCreate,
	AuditLogTypeAPIKeyRevoke,
	AuditLogTypeSecuritySettingsUpdate,
	AuditLogTypeRoutingChangeConfirm,
	AuditLogTypeRoutingChangeRevert,
	AuditLogTypeSettingCreate,
	AuditLogTypeSettingUpdate,
	AuditLogTypeSettingStatusUpdate,
	AuditLogTypeSettingDelete,
	AuditLogTypeProjectUpdate,
	AuditLogTypeProjectCreate,
	AuditLogTypeProjectDelete,
	AuditLogTypeAppUpdate,
	AuditLogTypeAppCreate,
	AuditLogTypeAppDelete,
	AuditLogTypeProjectEnvDelete,
}

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

	// AuditLogSourceAPIDelete is a delete endpoint.
	AuditLogSourceAPIDelete AuditLogSource = "api-delete"

	// AuditLogSourceAPIAction is an endpoint that makes something happen rather
	// than writing a record: a deploy, a restart, a stop, a cancel. Recording one
	// of those as api-update would say the stored configuration changed, which is
	// the opposite of what a reader needs to know about a restart.
	AuditLogSourceAPIAction AuditLogSource = "api-action"
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
