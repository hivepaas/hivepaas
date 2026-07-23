package apperrors

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

//nolint:err113
func NewErr(base error, s string) error {
	if base == nil {
		return errors.New(s)
	}
	return errors.Join(base, errors.New(s))
}

// Base errors
var (
	ErrInternal             = errors.New("ERR_INTERNAL")
	ErrBadRequest           = errors.New("ERR_BAD_REQUEST")
	ErrUnauthorized         = errors.New("ERR_UNAUTHORIZED")
	ErrForbidden            = errors.New("ERR_FORBIDDEN")
	ErrNotFound             = errors.New("ERR_NOT_FOUND")
	ErrConflict             = errors.New("ERR_CONFLICT")
	ErrPreconditionFailed   = errors.New("ERR_PRECONDITION_FAILED")
	ErrPreconditionRequired = errors.New("ERR_PRECONDITION_REQUIRED")
	ErrServiceUnavailable   = errors.New("ERR_SERVICE_UNAVAILABLE")
	ErrNotImplemented       = errors.New("ERR_NOT_IMPLEMENTED")
)

// errorStatusMap - mapping from base error to http status code
// Do not put non-base errors to this map
var errorStatusMap = map[error]int{
	ErrInternal:             http.StatusInternalServerError,
	ErrBadRequest:           http.StatusBadRequest,
	ErrUnauthorized:         http.StatusUnauthorized,
	ErrForbidden:            http.StatusForbidden,
	ErrNotFound:             http.StatusNotFound,
	ErrConflict:             http.StatusConflict,
	ErrPreconditionFailed:   http.StatusPreconditionFailed,
	ErrPreconditionRequired: http.StatusPreconditionRequired,
	ErrServiceUnavailable:   http.StatusServiceUnavailable,
	ErrNotImplemented:       http.StatusNotImplemented,
}

// errorWarnLevelMap defines the errors that are handled but unexpected to happen.
// Every error defined in this map will be notified at WARN level instead of ERROR.
var errorWarnLevelMap = map[error]bool{}

// grpcErrorStatusMap - mapping from HTTP status code to gRPC code
// Do not put non-base errors to this map
var grpcErrorStatusMap = map[error]codes.Code{
	ErrInternal:             codes.Unknown,
	ErrBadRequest:           codes.InvalidArgument,
	ErrUnauthorized:         codes.Unauthenticated,
	ErrForbidden:            codes.PermissionDenied,
	ErrNotFound:             codes.NotFound,
	ErrConflict:             codes.AlreadyExists,
	ErrPreconditionFailed:   codes.FailedPrecondition,
	ErrPreconditionRequired: codes.FailedPrecondition,
	ErrServiceUnavailable:   codes.Unavailable,
	ErrNotImplemented:       codes.Unimplemented,
}

// Popular errors
var (
	ErrPanic                    = NewErr(ErrInternal, "ERR_PANIC")
	ErrAlreadyExist             = NewErr(ErrConflict, "ERR_ALREADY_EXIST")
	ErrUnsupported              = NewErr(ErrPreconditionFailed, "ERR_UNSUPPORTED")
	ErrUnrecognized             = NewErr(ErrPreconditionFailed, "ERR_UNRECOGNIZED")
	ErrNonEditable              = NewErr(ErrPreconditionFailed, "ERR_NON_EDITABLE")
	ErrNonDeletable             = NewErr(ErrPreconditionFailed, "ERR_NON_DELETABLE")
	ErrInUse                    = NewErr(ErrConflict, "ERR_IN_USE")
	ErrInactive                 = NewErr(ErrPreconditionFailed, "ERR_INACTIVE")
	ErrMissing                  = NewErr(ErrPreconditionFailed, "ERR_MISSING")
	ErrParamMissing             = NewErr(ErrMissing, "ERR_PARAM_MISSING")
	ErrTooMany                  = NewErr(ErrPreconditionFailed, "ERR_TOO_MANY")
	ErrTooBig                   = NewErr(ErrPreconditionFailed, "ERR_TOO_BIG")
	ErrRequestTooBig            = NewErr(ErrTooBig, "ERR_REQUEST_TOO_BIG")
	ErrNotAllowed               = NewErr(ErrForbidden, "ERR_NOT_ALLOWED")
	ErrActionNotAllowed         = NewErr(ErrNotAllowed, "ERR_ACTION_NOT_ALLOWED")
	ErrActionNotAllowedByStatus = NewErr(ErrNotAllowed, "ERR_ACTION_NOT_ALLOWED_BY_STATUS")
	ErrActionNotAllowedByAdmin  = NewErr(ErrNotAllowed, "ERR_ACTION_NOT_ALLOWED_BY_ADMIN")
	ErrActionFailed             = NewErr(ErrPreconditionFailed, "ERR_ACTION_FAILED")
	ErrHTTPRequestFailed        = NewErr(ErrActionFailed, "ERR_HTTP_REQUEST_FAILED")
	ErrGRPCRequestFailed        = NewErr(ErrActionFailed, "ERR_GRPC_REQUEST_FAILED")
	ErrUnavailable              = NewErr(ErrPreconditionFailed, "ERR_UNAVAILABLE")
	ErrArgumentInvalid          = NewErr(ErrBadRequest, "ERR_ARGUMENT_INVALID")
	ErrValueInvalid             = NewErr(ErrPreconditionFailed, "ERR_VALUE_INVALID")
	ErrTokenInvalid             = NewErr(ErrValueInvalid, "ERR_TOKEN_INVALID")
	ErrMismatch                 = NewErr(ErrPreconditionFailed, "ERR_MISMATCH")
	ErrUpdateVerMismatched      = NewErr(ErrMismatch, "ERR_UPDATE_VER_MISMATCHED")
	ErrValidation               = NewErr(ErrBadRequest, "ERR_VALIDATION")
)

// Errors for session
var (
	ErrNoSession                   = NewErr(ErrUnauthorized, "ERR_NO_SESSION")
	ErrSessionJWTInvalid           = NewErr(ErrUnauthorized, "ERR_SESSION_JWT_INVALID")
	ErrSessionJWTExpired           = NewErr(ErrUnauthorized, "ERR_SESSION_JWT_EXPIRED")
	ErrSessionAPIKeyInvalid        = NewErr(ErrUnauthorized, "ERR_SESSION_API_KEY_INVALID")
	ErrSessionRefreshTokenRequired = NewErr(ErrUnauthorized, "ERR_SESSION_REFRESH_TOKEN_REQUIRED")
	ErrSSORequired                 = NewErr(ErrUnauthorized, "ERR_SSO_REQUIRED")
	ErrLoginInputInvalid           = NewErr(ErrUnauthorized, "ERR_LOGIN_INPUT_INVALID")
	ErrPasswordMismatched          = NewErr(ErrUnauthorized, "ERR_PASSWORD_MISMATCHED")
	ErrPasscodeMismatched          = NewErr(ErrUnauthorized, "ERR_PASSCODE_MISMATCHED")
	ErrTooManyLoginFailures        = NewErr(ErrUnauthorized, "ERR_TOO_MANY_LOGIN_FAILURES")
	ErrTooManyPasscodeAttempts     = NewErr(ErrUnauthorized, "ERR_TOO_MANY_PASSCODE_ATTEMPTS")
	ErrOAuthUserEmailNotReturned   = NewErr(ErrPreconditionFailed, "ERR_OAUTH_USER_EMAIL_NOT_RETURNED")
)

// Errors for user
var (
	ErrUserUnavailable             = NewErr(ErrUnauthorized, "ERR_USER_UNAVAILABLE")
	ErrUsernameUnavailable         = NewErr(ErrUnavailable, "ERR_USERNAME_UNAVAILABLE")
	ErrUserDemoUnauthorized        = NewErr(ErrUnauthorized, "ERR_USER_DEMO_UNAUTHORIZED")
	ErrUserStatusNotAllowAction    = NewErr(ErrNotAllowed, "ERR_USER_STATUS_NOT_ALLOW_ACTION")
	ErrUserAlreadySignUp           = NewErr(ErrPreconditionFailed, "ERR_USER_ALREADY_SIGN_UP")
	ErrUserNotCompleteMFASetup     = NewErr(ErrPreconditionFailed, "ERR_USER_NOT_COMPLETE_MFA_SETUP")
	ErrPasswordNotMeetRequirements = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_NOT_MEET_REQUIREMENTS")
	ErrPasswordResetTokenInvalid   = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_RESET_TOKEN_INVALID")
	ErrEmailUnavailable            = NewErr(ErrUnavailable, "ERR_EMAIL_UNAVAILABLE")
	ErrEmailChangeUnallowed        = NewErr(ErrNotAllowed, "ERR_EMAIL_CHANGE_UNALLOWED")
)

// Errors for api client
var (
	ErrAPIKeyInvalid = NewErr(ErrValueInvalid, "ERR_API_KEY_INVALID")
)

// Errors for settings
var (
	ErrUnconfigured                         = NewErr(ErrPreconditionFailed, "ERR_UNCONFIGURED")
	ErrSettingInactive                      = NewErr(ErrInactive, "ERR_SETTING_INACTIVE")
	ErrSettingNotFound                      = NewErr(ErrNotFound, "ERR_SETTING_NOT_FOUND")
	ErrSettingMissing                       = NewErr(ErrMissing, "ERR_SETTING_MISSING")
	ErrSettingViolation                     = NewErr(ErrForbidden, "ERR_SETTING_VIOLATION")
	ErrGlobalSettingRequired                = NewErr(ErrPreconditionRequired, "ERR_GLOBAL_SETTING_REQUIRED")
	ErrScopeSettingRequired                 = NewErr(ErrPreconditionRequired, "ERR_SCOPE_SETTING_REQUIRED")
	ErrObjectScopeInvalid                   = NewErr(ErrValueInvalid, "ERR_OBJECT_SCOPE_INVALID")
	ErrInheritedSettingNonUpdatable         = NewErr(ErrNonEditable, "ERR_INHERITED_SETTING_NON_UPDATABLE")
	ErrSettingTypeUnsupported               = NewErr(ErrUnsupported, "ERR_SETTING_TYPE_UNSUPPORTED")
	ErrEnvVarContainInvalidReference        = NewErr(ErrValueInvalid, "ERR_ENV_VAR_CONTAIN_INVALID_REFERENCE")
	ErrSharedEnvVarContainExternalReference = NewErr(ErrValueInvalid, "ERR_SHARED_ENV_VAR_CONTAIN_EXTERNAL_REFERENCE")
	ErrEnvVarExternalReferenceIsNotAllowed  = NewErr(ErrValueInvalid, "ERR_ENV_VAR_EXTERNAL_REFERENCE_IS_NOT_ALLOWED")
	ErrDomainInUse                          = NewErr(ErrInUse, "ERR_DOMAIN_IN_USE")
	ErrDomainUnallowed                      = NewErr(ErrSettingViolation, "ERR_DOMAIN_UNALLOWED")
	ErrSSLTypeUnsupported                   = NewErr(ErrUnsupported, "ERR_SSL_TYPE_UNSUPPORTED")
	ErrPrivateKeyTypeUnsupported            = NewErr(ErrUnsupported, "ERR_PRIVATE_KEY_TYPE_UNSUPPORTED")
	ErrAddressInvalid                       = NewErr(ErrValueInvalid, "ERR_ADDRESS_INVALID")
	ErrTokenTypeUnsupported                 = NewErr(ErrUnsupported, "ERR_TOKEN_TYPE_UNSUPPORTED")
	ErrWebhookTypeUnsupported               = NewErr(ErrUnsupported, "ERR_WEBHOOK_TYPE_UNSUPPORTED")
	ErrDataVerNewerThanSystemVer            = NewErr(ErrValueInvalid, "ERR_DATA_VER_NEWER_THAN_SYSTEM_VER")
)

// Errors for projects
var (
	ErrProjectNotFound           = NewErr(ErrNotFound, "ERR_PROJECT_NOT_FOUND")
	ErrProjectInactive           = NewErr(ErrInactive, "ERR_PROJECT_INACTIVE")
	ErrProjectNameNotAllowed     = NewErr(ErrNotAllowed, "ERR_PROJECT_NAME_NOT_ALLOWED")
	ErrProjectNetworkUnavailable = NewErr(ErrUnavailable, "ERR_PROJECT_NETWORK_UNAVAILABLE")
)

// Errors for apps
var (
	ErrAppNotFound                              = NewErr(ErrNotFound, "ERR_APP_NOT_FOUND")
	ErrAppInactive                              = NewErr(ErrInactive, "ERR_APP_INACTIVE")
	ErrMultiNodeClusterRequireRegistryForImages = NewErr(ErrPreconditionRequired, "ERR_MULTI_NODE_CLUSTER_REQUIRE_REGISTRY_FOR_IMAGES") //nolint:lll
	ErrDeploymentMethodRepoRequired             = NewErr(ErrUnconfigured, "ERR_DEPLOYMENT_METHOD_REPO_REQUIRED")
	ErrFeatureDisabled                          = NewErr(ErrInactive, "ERR_FEATURE_DISABLED")
)

// Errors for cluster
var (
	ErrServiceNotRunning             = NewErr(ErrServiceUnavailable, "ERR_SERVICE_NOT_RUNNING")
	ErrMountTypeUnsupported          = NewErr(ErrUnsupported, "ERR_MOUNT_TYPE_UNSUPPORTED")
	ErrVolumeAlreadyExists           = NewErr(ErrAlreadyExist, "ERR_VOLUME_ALREADY_EXISTS")
	ErrVolumeDriverUnsupported       = NewErr(ErrUnsupported, "ERR_VOLUME_DRIVER_UNSUPPORTED")
	ErrServiceModeReplicatedRequired = NewErr(ErrPreconditionFailed, "ERR_SERVICE_MODE_REPLICATED_REQUIRED")
)

// Errors for files
var (
	ErrFileScopeUnsupported        = NewErr(ErrUnsupported, "ERR_FILE_SCOPE_UNSUPPORTED")
	ErrFileSizeTooBig              = NewErr(ErrArgumentInvalid, "ERR_FILE_SIZE_TOO_BIG")
	ErrFileTypeNotSupported        = NewErr(ErrUnsupported, "ERR_FILE_TYPE_NOT_SUPPORTED")
	ErrFileExtNotSupported         = NewErr(ErrUnsupported, "ERR_FILE_EXT_NOT_SUPPORTED")
	ErrFileNameTooLong             = NewErr(ErrArgumentInvalid, "ERR_FILE_NAME_TOO_LONG")
	ErrFileTargetObjectUnavailable = NewErr(ErrArgumentInvalid, "ERR_FILE_TARGET_OBJECT_UNAVAILABLE")
	ErrArchiveFormatUnsupported    = NewErr(ErrUnsupported, "ERR_ARCHIVE_FORMAT_UNSUPPORTED")
	ErrEncryptionFormatUnsupported = NewErr(ErrUnsupported, "ERR_ENCRYPTION_FORMAT_UNSUPPORTED")
	ErrStorageTypeUnsupported      = NewErr(ErrUnsupported, "ERR_STORAGE_TYPE_UNSUPPORTED")
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
