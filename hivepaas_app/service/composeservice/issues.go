package composeservice

// Issue codes of reading a compose file, beside the import's own. Each is on
// the plan node of what it is about: an app's, or the env's for the file as a
// whole.
const (
	// Blocked: nothing is created until the file or the review says otherwise.
	CodeVariableRequired = "COMPOSE_VARIABLE_REQUIRED"
	CodeKeyConflict      = "COMPOSE_KEY_CONFLICT"
	CodeDependsCycle     = "COMPOSE_DEPENDS_CYCLE"
	CodeAppExists        = "COMPOSE_APP_EXISTS"

	// Skipped: the service is not created.
	CodeNoImage = "COMPOSE_NO_IMAGE"

	// Fixable: part of the service is left out, and the app is created without
	// it.
	CodeFileMissing       = "COMPOSE_FILE_MISSING"
	CodeMountDropped      = "COMPOSE_MOUNT_DROPPED"
	CodeCapabilityDropped = "COMPOSE_CAPABILITY_DROPPED"
	CodeDomainMissing     = "COMPOSE_DOMAIN_MISSING"
	CodeValueDropped      = "COMPOSE_VALUE_DROPPED"

	// Warnings: what is created may not do what the file meant.
	CodeNotSupported   = "COMPOSE_NOT_SUPPORTED"
	CodeBuildIgnored   = "COMPOSE_BUILD_IGNORED"
	CodeDirectoryEmpty = "COMPOSE_DIRECTORY_EMPTY"
	CodeVolumeExternal = "COMPOSE_VOLUME_EXTERNAL"
	CodeVolumeDriver   = "COMPOSE_VOLUME_DRIVER"
	CodeNetwork        = "COMPOSE_NETWORK"
	CodePlacement      = "COMPOSE_PLACEMENT"
	CodeVariableEmpty  = "COMPOSE_VARIABLE_EMPTY"
	CodeAliasTaken     = "COMPOSE_ALIAS_TAKEN"
	CodeSecretExists   = "COMPOSE_SECRET_EXISTS" //nolint:gosec // an issue code, not a credential

	// Notes: what the import does, needing no acceptance.
	CodeSecretVariable = "COMPOSE_SECRET_VARIABLE"
	CodeSecretWritten  = "COMPOSE_SECRET_WRITTEN" //nolint:gosec // an issue code, not a credential
	CodeSecretEnv      = "COMPOSE_SECRET_ENV"     //nolint:gosec // an issue code, not a credential
	CodeAliasAdded     = "COMPOSE_ALIAS_ADDED"
	CodeAppUsed        = "COMPOSE_APP_USED"
	CodeDirectoryFiles = "COMPOSE_DIRECTORY_FILES"
	CodeSettingRenamed = "COMPOSE_SETTING_RENAMED"
	CodeStartOrder     = "COMPOSE_START_ORDER"
	CodeLabelsDropped  = "COMPOSE_LABELS_DROPPED"
	CodeImageUnpinned  = "COMPOSE_IMAGE_UNPINNED"
	CodeJob            = "COMPOSE_JOB"
)
