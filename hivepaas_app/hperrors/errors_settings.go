package hperrors

// Errors for settings
var (
	ErrUnconfigured                         = NewErr(ErrPreconditionFailed, "ERR_UNCONFIGURED")
	ErrSettingNotFound                      = NewErr(ErrNotFound, "ERR_SETTING_NOT_FOUND")
	ErrSettingMissing                       = NewErr(ErrMissing, "ERR_SETTING_MISSING")
	ErrSettingViolation                     = NewErr(ErrForbidden, "ERR_SETTING_VIOLATION")
	ErrObjectScopeInvalid                   = NewErr(ErrValueInvalid, "ERR_OBJECT_SCOPE_INVALID")
	ErrInheritedSettingNonUpdatable         = NewErr(ErrNonEditable, "ERR_INHERITED_SETTING_NON_UPDATABLE")
	ErrSettingTypeUnsupported               = NewErr(ErrUnsupported, "ERR_SETTING_TYPE_UNSUPPORTED")
	ErrEnvVarCircularReference              = NewErr(ErrValueInvalid, "ERR_ENV_VAR_CIRCULAR_REFERENCE")
	ErrSharedEnvVarContainExternalReference = NewErr(ErrValueInvalid, "ERR_SHARED_ENV_VAR_CONTAIN_EXTERNAL_REFERENCE")
	ErrEnvVarExternalReferenceIsNotAllowed  = NewErr(ErrValueInvalid, "ERR_ENV_VAR_EXTERNAL_REFERENCE_IS_NOT_ALLOWED")
	ErrDomainInUse                          = NewErr(ErrInUse, "ERR_DOMAIN_IN_USE")
	ErrDomainUnallowed                      = NewErr(ErrSettingViolation, "ERR_DOMAIN_UNALLOWED")
	ErrSSLTypeUnsupported                   = NewErr(ErrUnsupported, "ERR_SSL_TYPE_UNSUPPORTED")
	ErrPrivateKeyTypeUnsupported            = NewErr(ErrUnsupported, "ERR_PRIVATE_KEY_TYPE_UNSUPPORTED")
	ErrAddressInvalid                       = NewErr(ErrValueInvalid, "ERR_ADDRESS_INVALID")
	ErrTokenTypeUnsupported                 = NewErr(ErrUnsupported, "ERR_TOKEN_TYPE_UNSUPPORTED")
	ErrWebhookTypeUnsupported               = NewErr(ErrUnsupported, "ERR_WEBHOOK_TYPE_UNSUPPORTED")
	ErrIMServiceUnsupported                 = NewErr(ErrUnsupported, "ERR_IM_SERVICE_UNSUPPORTED")
	ErrPasswordCurrentMismatched            = NewErr(ErrBadRequest, "ERR_PASSWORD_CURRENT_MISMATCHED")
	// ErrAppSecretMismatched is a wrong app secret, not a bad session: the caller
	// is still who they said they are, they just cannot prove they are the
	// operator. Deliberately a bad request rather than an unauthorized, so a
	// mistyped secret does not read as an expired login and log anyone out.
	ErrAppSecretMismatched = NewErr(ErrBadRequest, "ERR_APP_SECRET_MISMATCHED")
	// ErrTooManyAppSecretFailures spaces out guessing. Like the mismatch above it
	// is not an unauthorized: the session is fine, only this one check is being
	// made to wait, and answering 401 would log the operator out mid-change.
	ErrTooManyAppSecretFailures = NewErr(ErrTooMany, "ERR_TOO_MANY_APP_SECRET_FAILURES")

	// Routing changes that would leave HivePaaS unreachable. They are refusals of
	// the request, not reports of a broken system: nothing has happened yet, which
	// is the whole point of catching them here.
	ErrRoutingNoEnabledDomain = NewErr(ErrPreconditionFailed, "ERR_ROUTING_NO_ENABLED_DOMAIN")
	ErrRoutingCallerUnknown   = NewErr(ErrPreconditionFailed, "ERR_ROUTING_CALLER_UNKNOWN")
	ErrRoutingWouldLockYouOut = NewErr(ErrPreconditionFailed, "ERR_ROUTING_WOULD_LOCK_YOU_OUT")

	// Confirm-or-revert. A routing change is applied on probation and undone
	// unless the caller comes back through the new configuration and confirms it.
	ErrSettingsNoPendingChange  = NewErr(ErrNotFound, "ERR_SETTINGS_NO_PENDING_CHANGE")
	ErrSettingsChangeSuperseded = NewErr(ErrPreconditionFailed, "ERR_SETTINGS_CHANGE_SUPERSEDED")
	ErrSettingsConfirmTooEarly  = NewErr(ErrPreconditionFailed, "ERR_SETTINGS_CONFIRM_TOO_EARLY")

	// ErrSettingsChangeNotLive refuses to confirm a change that is no longer what
	// is running.
	//
	// Swarm undoes a service update whose task never becomes healthy - that is
	// what failure_action: rollback is for - and it puts the previous command back
	// without telling HivePaaS. The database then says one thing and the cluster
	// another. Accepting a confirmation there would end the trial and leave the
	// two disagreeing for good; refusing lets the deadline bring the database back
	// to what is actually running.
	ErrSettingsChangeNotLive = NewErr(ErrPreconditionFailed, "ERR_SETTINGS_CHANGE_NOT_LIVE")

	// ErrSettingInUse refuses to delete a setting something else still points at.
	//
	// A refusal rather than a cascade: the references live inside other settings'
	// payloads, so deleting this row would leave those payloads naming an id that
	// resolves to nothing - and the apply path treats a missing reference as fatal,
	// so each referencing object would fail at its next deploy, long after anybody
	// could connect the two events.
	ErrSettingInUse = NewErr(ErrPreconditionFailed, "ERR_SETTING_IN_USE")

	// ErrPlacementNoMatchingNode refuses placement rules no node satisfies.
	// Swarm's own answer to those is to leave every task Pending forever,
	// without an error anywhere an operator would look.
	ErrPlacementNoMatchingNode      = NewErr(ErrBadRequest, "ERR_PLACEMENT_NO_MATCHING_NODE")
	ErrPasswordNotMeetRequirements  = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_NOT_MEET_REQUIREMENTS")
	ErrPasswordHasWeakSequence      = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_HAS_WEAK_SEQUENCE")
	ErrPasswordTooSimilarToPrevious = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_TOO_SIMILAR_TO_PREVIOUS")
	ErrDataVerNewerThanSystemVer    = NewErr(ErrValueInvalid, "ERR_DATA_VER_NEWER_THAN_SYSTEM_VER")
)

// Errors for backup repositories
var (
	ErrBackupRepoStorageRequired  = NewErr(ErrBadRequest, "ERR_BACKUP_REPO_STORAGE_REQUIRED")
	ErrBackupRepoStorageAmbiguous = NewErr(ErrBadRequest, "ERR_BACKUP_REPO_STORAGE_AMBIGUOUS")
	ErrBackupRepoPasswordRequired = NewErr(ErrBadRequest, "ERR_BACKUP_REPO_PASSWORD_REQUIRED")
	// The repository lives on a node-local volume, so its location on the host must be resolvable
	// before kopia can be pointed at it.
	ErrBackupRepoVolumePathUnresolved = NewErr(ErrPreconditionFailed, "ERR_BACKUP_REPO_VOLUME_PATH_UNRESOLVED")
	ErrBackupRepoVolumeNodeRequired   = NewErr(ErrBadRequest, "ERR_BACKUP_REPO_VOLUME_NODE_REQUIRED")
	// The repository was re-encrypted but the new password could not be stored, and putting the
	// old one back failed too. Only a manual password change on the repository can fix it.
	ErrBackupRepoPasswordOutOfSync = NewErr(ErrPreconditionFailed, "ERR_BACKUP_REPO_PASSWORD_OUT_OF_SYNC")
	// The settings were saved but the repository would not take them, so backups keep running
	// with the previous compression and pack size until an update succeeds.
	ErrBackupRepoOptionsNotApplied = NewErr(ErrActionFailed, "ERR_BACKUP_REPO_OPTIONS_NOT_APPLIED")
	// A cleanup is already running for this repository. Pruning is slow and rewrites the
	// repository, so a second run is refused rather than queued behind the first.
	ErrBackupRepoCleanupInProgress = NewErr(ErrConflict, "ERR_BACKUP_REPO_CLEANUP_IN_PROGRESS")
)
