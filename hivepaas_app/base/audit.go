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

	// AuditLogTypeClusterUpdate records an operation carried out on the cluster
	// itself: a machine joined, the manager set changed, the build cache purged,
	// the database reconciled against Docker.
	//
	// Nodes, networks and volumes are stored as settings, so their creation and
	// removal already arrive as setting-create and setting-delete with ResType
	// naming which kind. This covers what is left: the operations that act on the
	// cluster rather than on its records. A node joining gets no type of its own
	// for that reason - it would be a third channel for node events, on top of
	// the two that already exist, which is harder to follow than a section.
	AuditLogTypeClusterUpdate AuditLogType = "cluster-update"

	// AuditLogTypeTaskCancel records work in flight being stopped by hand: a
	// deploy, an image build, a backup, a cleanup sweep.
	//
	// The only type tasks get, and it is about the person rather than the task.
	// Everything else in a task's life - queued, retried, failed, finished - is
	// the system acting on its own and already has the task row as its record.
	// A cancel is the one point where somebody reaches in and stops the work,
	// which is what is worth answering for later.
	//
	// One type across every scope, rather than folding the cancel into whatever
	// the task was doing. Tasks run under projects, environments, apps, users and
	// the install itself, and there is no type at all for some of those - the
	// entry's own scope says where it happened, which is what a type per scope
	// would have said at the cost of five more values in the type filter.
	AuditLogTypeTaskCancel AuditLogType = "task-cancel"

	// AuditLogTypeUserCreate, AuditLogTypeUserUpdate and AuditLogTypeUserDelete
	// record an account appearing, being changed, and being taken away.
	//
	// Split three ways for the reason projects and apps are: an account arriving
	// and an account disappearing are what somebody scanning the type column is
	// looking for, and folded into the edits a new admin would be
	// indistinguishable from a changed job title.
	//
	// One type for every edit, with section naming which form was used - account,
	// profile, password, password-reset, password-reset-request, mfa-setup,
	// mfa-remove. The two weightiest, a role change and a second factor being
	// taken off, are sections on the same reasoning as everything else in this
	// file; what moved is in the entry's detail, with the values, because a role
	// is not a secret and "member to admin" is the whole of what a reader wants.
	AuditLogTypeUserCreate AuditLogType = "user-create"
	AuditLogTypeUserUpdate AuditLogType = "user-update"
	AuditLogTypeUserDelete AuditLogType = "user-delete"

	// AuditLogTypeUserLogin records a session being handed out, and every attempt
	// that was turned away.
	//
	// The refusals are the half that exists nowhere else. A login that worked
	// leaves a session behind to find; a wrong password leaves nothing at all,
	// and a run of them against one account is the earliest notice anybody gets
	// that somebody is working at it. The backoff that slows those guesses down
	// counts them in redis, where the count is per account, expires in hours, and
	// is never read by a person.
	//
	// One type for every way in, with section naming which - password, passcode,
	// oauth, api-key, dev-mode. A type per method would put five values in the
	// filter to answer one question, and which door was used is the second
	// question a reader asks, not the first.
	//
	// A refused attempt names no actor. Who was on the other end is precisely
	// what could not be established; the account that was aimed at is recorded as
	// the entry's resource, and where the attempt came from is on every entry
	// anyway.
	AuditLogTypeUserLogin AuditLogType = "user-login"

	// AuditLogTypeUserLogout records a session being given up - the one in hand,
	// or every session the user has, which section tells apart.
	//
	// Worth recording for the same reason a login is: signing out everywhere is
	// what somebody does after losing a laptop, and also what somebody else does
	// after taking one. The two read identically in the moment and differently
	// afterwards, which is what a trail is for.
	AuditLogTypeUserLogout AuditLogType = "user-logout"

	// AuditLogTypeHivePaaSSecuritySettingsUpdate records a change to the
	// operator-level security switches, and the rotation of the app secret.
	//
	// One of the switches decides whether stored secrets may leave the server at
	// all, so the change is worth as much as the reveals it permits: without
	// this, a switch flipped on and back off leaves the reveals in between
	// looking like they were always allowed.
	//
	// The rotation is a section of this - app-secret-rotate, against
	// security-settings for the switches - rather than a type of its own. It is
	// the weightiest single action in here, because the secret wraps the data
	// encryption key and rotating it moves every stored secret in the system onto
	// a new key at once; but that is a reason to be able to find it, and (type,
	// section) is the indexed pair that finds it.
	//
	// Both sections record refusals as well as writes. An attempt at either from
	// an admin session is somebody working at the operator's own credential, and
	// that attempt exists nowhere else if it is not written down here.
	AuditLogTypeHivePaaSSecuritySettingsUpdate AuditLogType = "hivepaas-security-settings-update"

	// AuditLogTypeHivePaaSSettingsUpdateConfirm records somebody vouching that a
	// settings change left HivePaaS reachable, which is what stops it from being
	// undone.
	//
	// Not named after routing, which is what it used to say. The same trial
	// guards the HivePaaS service settings and traefik's startup command, and a
	// name that says routing makes two of the three read as something else -
	// exactly the drift this file's opening comment is about. Which page the
	// change was made on is the entry's section, the same value the change itself
	// carried.
	AuditLogTypeHivePaaSSettingsUpdateConfirm AuditLogType = "hivepaas-settings-update-confirm"

	// AuditLogTypeHivePaaSSettingsUpdateRevert records a settings change being
	// undone on request, before its deadline. The automatic undo is not recorded
	// here: it has no caller to attribute, and its record is the task row that
	// ran it.
	//
	// Confirm and revert stay types of their own rather than becoming sections of
	// hivepaas-settings-update, because section already carries which page was
	// written - routing, service, traefik-config - and one field cannot carry
	// both the page and the act without multiplying into one value per pair.
	// "What was taken back" is asked across every page at once, which is a
	// question only the type column can answer.
	AuditLogTypeHivePaaSSettingsUpdateRevert AuditLogType = "hivepaas-settings-update-revert"

	// AuditLogTypeHivePaaSSettingsUpdate records a change to HivePaaS's own
	// routing or service settings - the ones that decide whether HivePaaS is
	// reachable and how many replicas serve it - and to traefik's, which decide
	// the same thing for everything the install serves.
	//
	// Traefik shares the type rather than taking one of its own: its command and
	// its replica count are the install's wiring in exactly the sense the
	// HivePaaS pages are, and section - traefik-config, traefik-service - already
	// tells the four pages apart without adding values to the type filter.
	//
	// Separate from security-settings-update, which stays on its own type: that
	// one is about whether stored secrets may leave the server, which is a
	// different question from how the install is wired.
	AuditLogTypeHivePaaSSettingsUpdate AuditLogType = "hivepaas-settings-update"

	// AuditLogTypeHivePaaSAction records something done to the running install
	// rather than to its configuration: upgrading it to a new version, restarting
	// its services, telling them to re-read their config. Traefik's restart,
	// config reload and config reset are filed here too - cycling the ingress is
	// the same kind of act as cycling the app, and section - traefik-restart,
	// traefik-config-reload, traefik-config-reset - says which service it was.
	//
	// One type for the three, with section telling them apart. A version upgrade
	// is the weighty one and could have been given a type of its own, but section
	// is an indexed column that the listing filters on, so it now separates them
	// as well as a type would - and every type added is one more value in a
	// filter that has to stay readable.
	AuditLogTypeHivePaaSAction AuditLogType = "hivepaas-action"
)

var AllAuditLogTypes = []AuditLogType{
	AuditLogTypeSecretReveal,
	AuditLogTypeUserCreate,
	AuditLogTypeUserUpdate,
	AuditLogTypeUserDelete,
	AuditLogTypeUserLogin,
	AuditLogTypeUserLogout,
	AuditLogTypeAPIKeyCreate,
	AuditLogTypeAPIKeyRevoke,
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
	AuditLogTypeClusterUpdate,
	AuditLogTypeTaskCancel,
	AuditLogTypeHivePaaSSecuritySettingsUpdate,
	AuditLogTypeHivePaaSSettingsUpdateConfirm,
	AuditLogTypeHivePaaSSettingsUpdateRevert,
	AuditLogTypeHivePaaSSettingsUpdate,
	AuditLogTypeHivePaaSAction,
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
