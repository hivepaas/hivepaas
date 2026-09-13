package hperrors

var (
	ErrLoggingNotConfigured        = NewErr(ErrBadRequest, "ERR_LOGGING_NOT_CONFIGURED")
	ErrLoggingBackendNodeMissing   = NewErr(ErrBadRequest, "ERR_LOGGING_BACKEND_NODE_MISSING")
	ErrLoggingVolumeMissing        = NewErr(ErrBadRequest, "ERR_LOGGING_VOLUME_MISSING")
	ErrLoggingDeployFailed         = NewErr(ErrActionFailed, "ERR_LOGGING_DEPLOY_FAILED")
	ErrLoggingAPINetworkMissing    = NewErr(ErrActionFailed, "ERR_LOGGING_API_NETWORK_MISSING")
	ErrLoggingNotEnabled           = NewErr(ErrUnavailable, "ERR_LOGGING_NOT_ENABLED")
	ErrLoggingQueryEndpointMissing = NewErr(ErrUnavailable, "ERR_LOGGING_QUERY_ENDPOINT_MISSING")
)
