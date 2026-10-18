package hperrors

// Errors for projects
var (
	ErrProjectNotFound            = NewErr(ErrNotFound, "ERR_PROJECT_NOT_FOUND")
	ErrProjectEnvNotFound         = NewErr(ErrNotFound, "ERR_PROJECT_ENV_NOT_FOUND")
	ErrProjectInactive            = NewErr(ErrInactive, "ERR_PROJECT_INACTIVE")
	ErrProjectEnvInactive         = NewErr(ErrInactive, "ERR_PROJECT_ENV_INACTIVE")
	ErrProjectNameNotAllowed      = NewErr(ErrNotAllowed, "ERR_PROJECT_NAME_NOT_ALLOWED")
	ErrProjectNetworkUnavailable  = NewErr(ErrUnavailable, "ERR_PROJECT_NETWORK_UNAVAILABLE")
	ErrProjectEnvRemovalUnallowed = NewErr(ErrNotAllowed, "ERR_PROJECT_ENV_REMOVAL_UNALLOWED")
	// The project HivePaaS runs in, and what in it runs this installation: not
	// to be deleted, disabled or stopped by hand.
	ErrSystemProjectProtected    = NewErr(ErrNotAllowed, "ERR_SYSTEM_PROJECT_PROTECTED")
	ErrSystemProjectEnvProtected = NewErr(ErrNotAllowed, "ERR_SYSTEM_PROJECT_ENV_PROTECTED")
	ErrSystemAppProtected        = NewErr(ErrNotAllowed, "ERR_SYSTEM_APP_PROTECTED")
)

// Errors for apps
var (
	ErrAppNotFound                              = NewErr(ErrNotFound, "ERR_APP_NOT_FOUND")
	ErrAppInactive                              = NewErr(ErrInactive, "ERR_APP_INACTIVE")
	ErrAppIsCurrent                             = NewErr(ErrInactive, "ERR_APP_IS_CURRENT")
	ErrAppsNotInSameProjectEnv                  = NewErr(ErrNotAllowed, "ERR_APPS_NOT_IN_SAME_PROJECT_ENV")
	ErrAppServiceUnavailable                    = NewErr(ErrUnavailable, "ERR_APP_SERVICE_UNAVAILABLE")
	ErrAppCloneSettingsRequired                 = NewErr(ErrPreconditionRequired, "ERR_APP_CLONE_SETTING_REQUIRED")
	ErrMultiNodeClusterRequireRegistryForImages = NewErr(ErrPreconditionRequired, "ERR_MULTI_NODE_CLUSTER_REQUIRE_REGISTRY_FOR_IMAGES") //nolint:lll
	ErrDeploymentMethodRepoRequired             = NewErr(ErrUnconfigured, "ERR_DEPLOYMENT_METHOD_REPO_REQUIRED")
	ErrFeatureDisabled                          = NewErr(ErrInactive, "ERR_FEATURE_DISABLED")
	// A build variable that uses a secret, which the Dockerfile declares with ARG:
	// a build argument's value is written into the image's history.
	ErrBuildSecretDeclaredAsArg = NewErr(ErrNotAllowed, "ERR_BUILD_SECRET_DECLARED_AS_ARG")
	// A function whose runtime the running release names no image for.
	ErrFunctionRuntimeUnavailable = NewErr(ErrUnsupported, "ERR_FUNCTION_RUNTIME_UNAVAILABLE")
	// A function is deployed from its code, and only a function is: the
	// deployment method follows the app's kind.
	ErrDeploymentMethodFunctionRequired  = NewErr(ErrPreconditionFailed, "ERR_DEPLOYMENT_METHOD_FUNCTION_REQUIRED")
	ErrDeploymentMethodFunctionUnallowed = NewErr(ErrPreconditionFailed, "ERR_DEPLOYMENT_METHOD_FUNCTION_UNALLOWED")
	// An app is a function from its creation, and stays one.
	ErrAppKindFunctionUnchangeable = NewErr(ErrNonEditable, "ERR_APP_KIND_FUNCTION_UNCHANGEABLE")
	// ErrAppManagedFieldLocked refuses an edit to what an app's managed engine
	// owns, naming the field. It is changed through the engine's settings, or the
	// app unmanaged first.
	ErrAppManagedFieldLocked = NewErr(ErrNonEditable, "ERR_APP_MANAGED_FIELD_LOCKED")
	// ErrAppManagedVolumeNotPinned refuses a managed app whose data volume is not
	// pinned to a node: its archive and its artifacts are decided by that node.
	ErrAppManagedVolumeNotPinned = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_VOLUME_NOT_PINNED")
	// ErrAppManagedArchiveUnreachable refuses a volume backup repository the
	// database's node cannot mount: one pinned to another node.
	ErrAppManagedArchiveUnreachable = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_ARCHIVE_UNREACHABLE")
	// ErrAppManagedVerifierNeedsObjectStorage refuses a verifier app for a
	// database archiving to a volume: a verifier reads the archive from where any
	// node can, which a volume is not.
	ErrAppManagedVerifierNeedsObjectStorage = NewErr(ErrPreconditionFailed,
		"ERR_APP_MANAGED_VERIFIER_NEEDS_OBJECT_STORAGE")
	// ErrAppManagedVerifierExists refuses a second verifier app: the restore test
	// moves back into the database's container first, which deletes the one there is.
	ErrAppManagedVerifierExists = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_VERIFIER_EXISTS")
	// ErrAppManagedRestoreTargetInvalid refuses a restore target that does not
	// say what to recover to.
	ErrAppManagedRestoreTargetInvalid = NewErr(ErrArgumentInvalid, "ERR_APP_MANAGED_RESTORE_TARGET_INVALID")
	// ErrAppManagedRestoreOutsideWindow refuses a restore target the archive does
	// not reach: before its oldest base backup, after its last archived WAL, or
	// with no base backup yet.
	ErrAppManagedRestoreOutsideWindow = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_RESTORE_OUTSIDE_WINDOW")
	// ErrAppManagedRestoreNoBackup refuses a restore of a database whose first
	// base backup is not taken yet.
	ErrAppManagedRestoreNoBackup = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_RESTORE_NO_BACKUP")
	// ErrAppManagedRestoreFailed ends a restore that cannot go on: the recovery
	// failed, or the first base backup did. What went wrong is in the task's log.
	ErrAppManagedRestoreFailed = NewErr(ErrActionFailed, "ERR_APP_MANAGED_RESTORE_FAILED")
	// ErrAppManagedRestorePointNotArchived refuses a restore point whose WAL was
	// not archived in time: it could not be restored to yet, so it is not
	// recorded. Archiving may be failing.
	ErrAppManagedRestorePointNotArchived = NewErr(ErrActionFailed, "ERR_APP_MANAGED_RESTORE_POINT_NOT_ARCHIVED")
	// ErrAppDeployRestorePointFailed is a deployment that asks for a restore
	// point on a managed database first and could not have it made: it stops
	// before anything changes.
	ErrAppDeployRestorePointFailed = NewErr(ErrActionFailed, "ERR_APP_DEPLOY_RESTORE_POINT_FAILED")
	// ErrAppManagedSwitchNotRestored refuses switching consumers to a database
	// that was not restored from another: there is no database they would switch
	// from.
	ErrAppManagedSwitchNotRestored = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_SWITCH_NOT_RESTORED")
	// ErrAppManagedSwitchNotReady refuses switching consumers to a database still
	// being restored.
	ErrAppManagedSwitchNotReady = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_SWITCH_NOT_READY")
	// ErrAppManagedSwitchSourceGone refuses a switch from a source no longer in
	// the environment: the key its consumers name cannot be read.
	ErrAppManagedSwitchSourceGone = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_SWITCH_SOURCE_GONE")
	// ErrAppManagedReplicaUnsupported refuses a replica of a database whose
	// scripts predate replicas: its own would start empty, not as a standby.
	ErrAppManagedReplicaUnsupported = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_REPLICA_UNSUPPORTED")
	// ErrAppManagedReplicaNeedsObjectStorage refuses a replica of a database whose
	// archive is on a volume: the replica is seeded from the archive, on its own
	// node.
	ErrAppManagedReplicaNeedsObjectStorage = NewErr(ErrPreconditionFailed,
		"ERR_APP_MANAGED_REPLICA_NEEDS_OBJECT_STORAGE")
	// ErrAppManagedPromoteUnsupported refuses a promotion in a cluster whose
	// scripts predate it: they could not fence the old primary.
	ErrAppManagedPromoteUnsupported = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_PROMOTE_UNSUPPORTED")
	// ErrAppManagedPromoteFailed is a promotion that did not happen: the fence
	// was lifted, or is kept when whether the member was promoted is not known.
	ErrAppManagedPromoteFailed = NewErr(ErrActionFailed, "ERR_APP_MANAGED_PROMOTE_FAILED")
	// ErrAppManagedClusterHeld refuses to delete the app a replica was promoted
	// over: it holds the cluster, its new primary among its members.
	ErrAppManagedClusterHeld = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_CLUSTER_HELD")
	// ErrAppManagedIsWriter refuses to remove the cluster's primary.
	ErrAppManagedIsWriter = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_IS_WRITER")
	// ErrAppManagedIsMember refuses on a member what only its primary does.
	ErrAppManagedIsMember = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_IS_MEMBER")
	// ErrAppManagedHasMembers refuses to unmanage a primary while it has members.
	ErrAppManagedHasMembers = NewErr(ErrPreconditionFailed, "ERR_APP_MANAGED_HAS_MEMBERS")
	// What only a function has: a test run, say.
	ErrAppNotFunction = NewErr(ErrPreconditionFailed, "ERR_APP_NOT_FUNCTION")
)

// Errors for autoscale
var (
	// What the app scales on cannot be read now: it is not turned on.
	ErrAutoscaleUnreadable = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_UNREADABLE")
	// An app other than a function scales on its requests, its CPU or both.
	ErrAutoscaleNoSignal = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_NO_SIGNAL")
	// An app publishing a port in host mode runs one replica a node at most.
	ErrAutoscaleHostPorts = NewErr(ErrPreconditionFailed, "ERR_AUTOSCALE_HOST_PORTS")
)

// Errors for sources
var (
	ErrRepoNotFound             = NewErr(ErrNotFound, "ERR_REPO_NOT_FOUND")
	ErrRepoTypeUnsupported      = NewErr(ErrUnsupported, "ERR_REPO_TYPE_UNSUPPORTED")
	ErrRepoRefNotFound          = NewErr(ErrNotFound, "ERR_REPO_REF_NOT_FOUND")
	ErrPullRequestNotFound      = NewErr(ErrNotFound, "ERR_PULL_REQUEST_NOT_FOUND")
	ErrPullRequestInvalid       = NewErr(ErrValueInvalid, "ERR_PULL_REQUEST_INVALID")
	ErrGitTypeUnsupported       = NewErr(ErrUnsupported, "ERR_GIT_TYPE_UNSUPPORTED")
	ErrGitAuthMethodUnsupported = NewErr(ErrUnsupported, "ERR_GIT_AUTH_METHOD_UNSUPPORTED")
	ErrGitLogOutputUnexpected   = NewErr(ErrPreconditionFailed, "ERR_GIT_LOG_OUTPUT_UNEXPECTED")
)
