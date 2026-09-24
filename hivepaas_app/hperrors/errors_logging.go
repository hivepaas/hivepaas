package hperrors

var (
	ErrLoggingNotConfigured        = NewErr(ErrBadRequest, "ERR_LOGGING_NOT_CONFIGURED")
	ErrLoggingVolumeMissing        = NewErr(ErrBadRequest, "ERR_LOGGING_VOLUME_MISSING")
	ErrLoggingAPINetworkMissing    = NewErr(ErrActionFailed, "ERR_LOGGING_API_NETWORK_MISSING")
	ErrLoggingNotEnabled           = NewErr(ErrUnavailable, "ERR_LOGGING_NOT_ENABLED")
	ErrLoggingQueryEndpointMissing = NewErr(ErrUnavailable, "ERR_LOGGING_QUERY_ENDPOINT_MISSING")

	ErrLoggingSettingsInvalid = NewErr(ErrValueInvalid, "ERR_LOGGING_SETTINGS_INVALID")
	ErrLoggingVolumeImmutable = NewErr(ErrValueInvalid, "ERR_LOGGING_VOLUME_IMMUTABLE")
	// ErrLoggingAppStillRunning is a save that would take an app of the stack
	// down without saying so: switching logging off, or handing the backend or
	// the collector to a system HivePaaS does not run.
	ErrLoggingAppStillRunning = NewErr(ErrPreconditionFailed, "ERR_LOGGING_APP_STILL_RUNNING")
)
