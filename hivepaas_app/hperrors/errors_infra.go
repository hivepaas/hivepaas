package hperrors

// Errors for cluster
var (
	ErrNodeWithLabelNotAvailable     = NewErr(ErrUnavailable, "ERR_NODE_WITH_LABEL_NOT_AVAILABLE")
	ErrServiceNotRunning             = NewErr(ErrServiceUnavailable, "ERR_SERVICE_NOT_RUNNING")
	ErrServiceModeReplicatedRequired = NewErr(ErrPreconditionFailed, "ERR_SERVICE_MODE_REPLICATED_REQUIRED")
	ErrActiveContainerNotFound       = NewErr(ErrNotFound, "ERR_ACTIVE_CONTAINER_NOT_FOUND")
	// Swarm forbids changing the service mode variant in place, so HivePaaS has to recreate the
	// service. That is refused while the app is stopped: a stopped app is a replicated service
	// scaled to 0, and global mode has no such notion, so converting it would silently start the
	// app on every node.
	ErrServiceModeChangeRequiresRunningApp = NewErr(ErrPreconditionFailed,
		"ERR_SERVICE_MODE_CHANGE_REQUIRES_RUNNING_APP")
)

// Errors for files
var (
	ErrDirNotCreated               = NewErr(ErrActionFailed, "ERR_DIR_NOT_CREATED")
	ErrFileScopeUnsupported        = NewErr(ErrUnsupported, "ERR_FILE_SCOPE_UNSUPPORTED")
	ErrFileSizeTooBig              = NewErr(ErrArgumentInvalid, "ERR_FILE_SIZE_TOO_BIG")
	ErrFileTypeNotSupported        = NewErr(ErrUnsupported, "ERR_FILE_TYPE_NOT_SUPPORTED")
	ErrFileExtNotSupported         = NewErr(ErrUnsupported, "ERR_FILE_EXT_NOT_SUPPORTED")
	ErrFileNameTooLong             = NewErr(ErrArgumentInvalid, "ERR_FILE_NAME_TOO_LONG")
	ErrArchiveFormatUnsupported    = NewErr(ErrUnsupported, "ERR_ARCHIVE_FORMAT_UNSUPPORTED")
	ErrEncryptionFormatUnsupported = NewErr(ErrUnsupported, "ERR_ENCRYPTION_FORMAT_UNSUPPORTED")
	ErrStorageTypeUnsupported      = NewErr(ErrUnsupported, "ERR_STORAGE_TYPE_UNSUPPORTED")
)

// nolint Errors from infrastructure
var (
	ErrInfra                   = NewErr(ErrInternal, "ERR_INFRA")
	ErrInfraUnknown            = NewErr(ErrInternal, "ERR_INFRA_UNKNOWN")
	ErrInfraInternal           = NewErr(ErrInternal, "ERR_INFRA_INTERNAL")
	ErrInfraActionFailed       = NewErr(ErrActionFailed, "ERR_INFRA_ACTION_FAILED")
	ErrInfraInvalidArgument    = NewErr(ErrArgumentInvalid, "ERR_INFRA_INVALID_ARGUMENT")
	ErrInfraNotFound           = NewErr(ErrNotFound, "ERR_INFRA_NOT_FOUND")
	ErrInfraAlreadyExists      = NewErr(ErrAlreadyExist, "ERR_INFRA_ALREADY_EXISTS")
	ErrInfraPermissionDenied   = NewErr(ErrUnauthorized, "ERR_INFRA_PERMISSION_DENIED")
	ErrInfraResourceExhausted  = NewErr(ErrPreconditionFailed, "ERR_INFRA_RESOURCE_EXHAUSTED")
	ErrInfraFailedPrecondition = NewErr(ErrPreconditionFailed, "ERR_INFRA_FAILED_PRECONDITION")
	ErrInfraConflict           = NewErr(ErrConflict, "ERR_INFRA_CONFLICT")
	ErrInfraNotModified        = NewErr(ErrPreconditionFailed, "ERR_INFRA_NOT_MODIFIED")
	ErrInfraAborted            = NewErr(ErrPreconditionFailed, "ERR_INFRA_ABORTED")
	ErrInfraOutOfRange         = NewErr(ErrPreconditionFailed, "ERR_INFRA_OUT_OF_RANGE")
	ErrInfraNotImplemented     = NewErr(ErrNotImplemented, "ERR_INFRA_NOT_IMPLEMENTED")
	ErrInfraUnavailable        = NewErr(ErrUnavailable, "ERR_INFRA_UNAVAILABLE")
	ErrInfraDataLoss           = NewErr(ErrPreconditionFailed, "ERR_INFRA_DATA_LOSS")
	ErrInfraUnauthorized       = NewErr(ErrUnauthorized, "ERR_INFRA_UNAUTHORIZED")
)
