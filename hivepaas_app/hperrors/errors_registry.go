package hperrors

// Errors for reading an image registry
var (
	ErrRegistryUnavailable = NewErr(ErrUnavailable, "ERR_REGISTRY_UNAVAILABLE")
	ErrRegistryRateLimited = NewErr(ErrUnavailable, "ERR_REGISTRY_RATE_LIMITED")
)

// Errors for the registry HivePaaS provisions and runs for itself
var (
	ErrRegistryNotConfigured    = NewErr(ErrPreconditionFailed, "ERR_REGISTRY_NOT_CONFIGURED")
	ErrRegistrySettingsInvalid  = NewErr(ErrValueInvalid, "ERR_REGISTRY_SETTINGS_INVALID")
	ErrRegistryStorageImmutable = NewErr(ErrValueInvalid, "ERR_REGISTRY_STORAGE_IMMUTABLE")
	ErrRegistryUnreachable      = NewErr(ErrUnavailable, "ERR_REGISTRY_UNREACHABLE")
	// ErrRegistryAppStillRunning refuses a save that would switch the registry off
	// while its app is still there. Switching off used to leave the app behind
	// with no screen to remove it from, so taking it down is now part of the same
	// save - and taking something down is not something a save may do unasked.
	ErrRegistryAppStillRunning = NewErr(ErrPreconditionFailed, "ERR_REGISTRY_APP_STILL_RUNNING")
)
