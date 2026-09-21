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
)
