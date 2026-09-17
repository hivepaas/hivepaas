package hperrors

// Errors for reading an image registry
var (
	ErrRegistryUnavailable = NewErr(ErrUnavailable, "ERR_REGISTRY_UNAVAILABLE")
	ErrRegistryRateLimited = NewErr(ErrUnavailable, "ERR_REGISTRY_RATE_LIMITED")
)
