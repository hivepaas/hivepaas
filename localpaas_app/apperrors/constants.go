package apperrors

import (
	"errors"
	"net/http"
)

// Base errors with equivalent http status code
var (
	ErrInternalServer = errors.New("ERR_INTERNAL_SERVER")
	ErrBadRequest     = errors.New("ERR_BAD_REQUEST")
	ErrUnauthorized   = errors.New("ERR_UNAUTHORIZED")
	ErrForbidden      = errors.New("ERR_FORBIDDEN")
	ErrNotFound       = errors.New("ERR_NOT_FOUND")
	ErrConflict       = errors.New("ERR_CONFLICT")
	ErrNotImplemented = errors.New("ERR_NOT_IMPLEMENTED")

	ErrPanic                = errors.New("ERR_PANIC")
	ErrAlreadyExist         = errors.New("ERR_ALREADY_EXIST")
	ErrUnsupported          = errors.New("ERR_UNSUPPORTED")
	ErrNonEditable          = errors.New("ERR_NON_EDITABLE")
	ErrNonDeletable         = errors.New("ERR_NON_DELETABLE")
	ErrResourceInUse        = errors.New("ERR_RESOURCE_IN_USE")
	ErrResourceInactive     = errors.New("ERR_RESOURCE_INACTIVE")
	ErrResourceMissing      = errors.New("ERR_RESOURCE_MISSING")
	ErrTooMany              = errors.New("ERR_TOO_MANY")
	ErrActionNotAllowed     = errors.New("ERR_ACTION_NOT_ALLOWED")
	ErrActionFailed         = errors.New("ERR_ACTION_FAILED")
	ErrUnavailable          = errors.New("ERR_UNAVAILABLE")
	ErrParamInvalid         = errors.New("ERR_PARAM_INVALID")
	ErrTypeInvalid          = errors.New("ERR_TYPE_INVALID")
	ErrValueInvalid         = errors.New("ERR_VALUE_INVALID")
	ErrTokenInvalid         = errors.New("ERR_TOKEN_INVALID")
	ErrUpdateVerMismatched  = errors.New("ERR_UPDATE_VER_MISMATCHED")
	ErrStatusNotAllowAction = errors.New("ERR_STATUS_NOT_ALLOW_ACTION")
	ErrURLInvalid           = errors.New("ERR_URL_INVALID")
	ErrValidation           = errors.New("ERR_VALIDATION")
)

// Errors for session
var (
	ErrNoSession                   = errors.New("ERR_NO_SESSION")
	ErrSessionJWTInvalid           = errors.New("ERR_SESSION_JWT_INVALID")
	ErrSessionJWTExpired           = errors.New("ERR_SESSION_JWT_EXPIRED")
	ErrSessionAPIKeyInvalid        = errors.New("ERR_SESSION_API_KEY_INVALID")
	ErrSessionRefreshTokenRequired = errors.New("ERR_SESSION_REFRESH_TOKEN_REQUIRED")
	ErrSSORequired                 = errors.New("ERR_SSO_REQUIRED")
	ErrLoginInputInvalid           = errors.New("ERR_LOGIN_INPUT_INVALID")
	ErrPasswordMismatched          = errors.New("ERR_PASSWORD_MISMATCHED")
	ErrPasscodeMismatched          = errors.New("ERR_PASSCODE_MISMATCHED")
	ErrTooManyLoginFailures        = errors.New("ERR_TOO_MANY_LOGIN_FAILURES")
	ErrTooManyPasscodeAttempts     = errors.New("ERR_TOO_MANY_PASSCODE_ATTEMPTS")
)

// Errors for user
var (
	ErrUserUnavailable             = errors.New("ERR_USER_UNAVAILABLE")
	ErrUserStatusNotAllowAction    = errors.New("ERR_USER_STATUS_NOT_ALLOW_ACTION")
	ErrUserAlreadySignUp           = errors.New("ERR_USER_ALREADY_SIGN_UP")
	ErrUserNotCompleteMFASetup     = errors.New("ERR_USER_NOT_COMPLETE_MFA_SETUP")
	ErrPasswordNotMeetRequirements = errors.New("ERR_PASSWORD_NOT_MEET_REQUIREMENTS")
	ErrPasswordResetTokenInvalid   = errors.New("ERR_PASSWORD_RESET_TOKEN_INVALID")
	ErrEmailUnavailable            = errors.New("ERR_EMAIL_UNAVAILABLE")
	ErrEmailChangeUnallowed        = errors.New("ERR_EMAIL_CHANGE_UNALLOWED")
)

// Errors for api client
var (
	ErrAPIKeyMismatched = errors.New("ERR_API_KEY_MISMATCHED")
	ErrAPIKeyInvalid    = errors.New("ERR_API_KEY_INVALID")
)

// Errors for settings
var (
	ErrUnconfigured                 = errors.New("ERR_UNCONFIGURED")
	ErrSettingInactive              = errors.New("ERR_SETTING_INACTIVE")
	ErrSettingMissing               = errors.New("ERR_SETTING_MISSING")
	ErrSettingViolated              = errors.New("ERR_SETTING_VIOLATED")
	ErrSettingUnallowed             = errors.New("ERR_SETTING_UNALLOWED")
	ErrGlobalSettingRequired        = errors.New("ERR_GLOBAL_SETTING_REQUIRED")
	ErrScopeSettingRequired         = errors.New("ERR_SCOPE_SETTING_REQUIRED")
	ErrInheritedSettingNonUpdatable = errors.New("ERR_INHERITED_SETTING_NON_UPDATABLE")
	ErrSettingTypeInvalid           = errors.New("ERR_SETTING_TYPE_INVALID")
	ErrDataVerNewerThanSystemVer    = errors.New("ERR_DATA_VER_NEWER_THAN_SYSTEM_VER")
)

// Errors for projects
var (
	ErrProjectInactive = errors.New("ERR_PROJECT_INACTIVE")
)

// Errors for apps
var (
	ErrAppInactive                              = errors.New("ERR_APP_INACTIVE")
	ErrMultiNodeClusterRequireRegistryForImages = errors.New("ERR_MULTI_NODE_CLUSTER_REQUIRE_REGISTRY_FOR_IMAGES")
)

// Errors for cluster
var (
	ErrNodeRequiredByLocalPaaSApp = errors.New("ERR_NODE_REQUIRED_BY_LOCALPAAS_APP")
	ErrServiceNotRunning          = errors.New("ERR_SERVICE_NOT_RUNNING")
)

// nolint Errors from infrastructure
var (
	ErrInfra                   = errors.New("ERR_INFRA")
	ErrInfraUnknown            = errors.New("ERR_INFRA_UNKNOWN")
	ErrInfraInternal           = errors.Join(errors.New("ERR_INFRA_INTERNAL"), ErrInternalServer)
	ErrInfraActionFailed       = errors.New("ERR_INFRA_ACTION_FAILED")
	ErrInfraInvalidArgument    = errors.Join(errors.New("ERR_INFRA_INVALID_ARGUMENT"), ErrParamInvalid)
	ErrInfraNotFound           = errors.Join(errors.New("ERR_INFRA_NOT_FOUND"), ErrNotFound)
	ErrInfraAlreadyExists      = errors.Join(errors.New("ERR_INFRA_ALREADY_EXISTS"), ErrAlreadyExist)
	ErrInfraPermissionDenied   = errors.New("ERR_INFRA_PERMISSION_DENIED")
	ErrInfraResourceExhausted  = errors.New("ERR_INFRA_RESOURCE_EXHAUSTED")
	ErrInfraFailedPrecondition = errors.New("ERR_INFRA_FAILED_PRECONDITION")
	ErrInfraConflict           = errors.Join(errors.New("ERR_INFRA_CONFLICT"), ErrConflict)
	ErrInfraNotModified        = errors.New("ERR_INFRA_NOT_MODIFIED")
	ErrInfraAborted            = errors.New("ERR_INFRA_ABORTED")
	ErrInfraOutOfRange         = errors.New("ERR_INFRA_OUT_OF_RANGE")
	ErrInfraNotImplemented     = errors.Join(errors.New("ERR_INFRA_NOT_IMPLEMENTED"), ErrNotImplemented)
	ErrInfraUnavailable        = errors.Join(errors.New("ERR_INFRA_UNAVAILABLE"), ErrUnavailable)
	ErrInfraDataLoss           = errors.New("ERR_INFRA_DATA_LOSS")
	ErrInfraUnauthorized       = errors.Join(errors.New("ERR_INFRA_UNAUTHORIZED"), ErrUnauthorized)
)

// errorStatusMap - mapping from error to http status code
var errorStatusMap = map[error]int{
	// Base errors
	ErrInternalServer: http.StatusInternalServerError,
	ErrBadRequest:     http.StatusBadRequest,
	ErrUnauthorized:   http.StatusUnauthorized,
	ErrForbidden:      http.StatusForbidden,
	ErrNotFound:       http.StatusNotFound,
	ErrConflict:       http.StatusConflict,
	ErrNotImplemented: http.StatusNotImplemented,

	ErrPanic:                http.StatusInternalServerError,
	ErrAlreadyExist:         http.StatusConflict,
	ErrUnsupported:          http.StatusUnprocessableEntity,
	ErrNonEditable:          http.StatusUnprocessableEntity,
	ErrNonDeletable:         http.StatusUnprocessableEntity,
	ErrResourceInUse:        http.StatusUnprocessableEntity,
	ErrResourceInactive:     http.StatusUnprocessableEntity,
	ErrResourceMissing:      http.StatusUnprocessableEntity,
	ErrTooMany:              http.StatusUnprocessableEntity,
	ErrActionNotAllowed:     http.StatusUnprocessableEntity,
	ErrActionFailed:         http.StatusUnprocessableEntity,
	ErrUnavailable:          http.StatusUnprocessableEntity,
	ErrParamInvalid:         http.StatusUnprocessableEntity,
	ErrTypeInvalid:          http.StatusUnprocessableEntity,
	ErrValueInvalid:         http.StatusUnprocessableEntity,
	ErrTokenInvalid:         http.StatusUnprocessableEntity,
	ErrUpdateVerMismatched:  http.StatusUnprocessableEntity,
	ErrStatusNotAllowAction: http.StatusUnprocessableEntity,
	ErrURLInvalid:           http.StatusUnprocessableEntity,
	ErrValidation:           http.StatusBadRequest,

	// Session errors
	ErrNoSession:                   http.StatusUnauthorized,
	ErrSessionJWTInvalid:           http.StatusUnauthorized,
	ErrSessionJWTExpired:           http.StatusUnauthorized,
	ErrSessionAPIKeyInvalid:        http.StatusUnauthorized,
	ErrSessionRefreshTokenRequired: http.StatusUnauthorized,
	ErrSSORequired:                 http.StatusUnauthorized,
	ErrLoginInputInvalid:           http.StatusUnauthorized,
	ErrPasswordMismatched:          http.StatusUnauthorized,
	ErrPasscodeMismatched:          http.StatusUnauthorized,
	ErrTooManyLoginFailures:        http.StatusUnauthorized,
	ErrTooManyPasscodeAttempts:     http.StatusUnauthorized,

	// User errors
	ErrUserUnavailable:             http.StatusUnprocessableEntity,
	ErrUserStatusNotAllowAction:    http.StatusUnprocessableEntity,
	ErrUserAlreadySignUp:           http.StatusUnprocessableEntity,
	ErrUserNotCompleteMFASetup:     http.StatusUnprocessableEntity,
	ErrPasswordNotMeetRequirements: http.StatusUnprocessableEntity,
	ErrPasswordResetTokenInvalid:   http.StatusUnprocessableEntity,
	ErrEmailUnavailable:            http.StatusUnprocessableEntity,
	ErrEmailChangeUnallowed:        http.StatusUnprocessableEntity,

	// Api client errors
	ErrAPIKeyMismatched: http.StatusUnauthorized,
	ErrAPIKeyInvalid:    http.StatusUnauthorized,

	// Settings errors
	ErrUnconfigured:                 http.StatusUnprocessableEntity,
	ErrSettingInactive:              http.StatusUnprocessableEntity,
	ErrSettingMissing:               http.StatusUnprocessableEntity,
	ErrSettingViolated:              http.StatusUnprocessableEntity,
	ErrSettingUnallowed:             http.StatusUnprocessableEntity,
	ErrGlobalSettingRequired:        http.StatusUnprocessableEntity,
	ErrScopeSettingRequired:         http.StatusUnprocessableEntity,
	ErrInheritedSettingNonUpdatable: http.StatusUnprocessableEntity,
	ErrSettingTypeInvalid:           http.StatusUnprocessableEntity,
	ErrDataVerNewerThanSystemVer:    http.StatusUnprocessableEntity,

	// Project errors
	ErrProjectInactive: http.StatusUnprocessableEntity,

	// App errors
	ErrAppInactive: http.StatusUnprocessableEntity,
	ErrMultiNodeClusterRequireRegistryForImages: http.StatusUnprocessableEntity,

	// Cluster errors
	ErrNodeRequiredByLocalPaaSApp: http.StatusUnprocessableEntity,
	ErrServiceNotRunning:          http.StatusUnprocessableEntity,

	// Errors from infrastructure
	ErrInfra:                   http.StatusInternalServerError,
	ErrInfraUnknown:            http.StatusInternalServerError,
	ErrInfraInternal:           http.StatusInternalServerError,
	ErrInfraActionFailed:       http.StatusUnprocessableEntity,
	ErrInfraInvalidArgument:    http.StatusBadRequest,
	ErrInfraNotFound:           http.StatusNotFound,
	ErrInfraAlreadyExists:      http.StatusConflict,
	ErrInfraPermissionDenied:   http.StatusUnauthorized,
	ErrInfraResourceExhausted:  http.StatusUnprocessableEntity,
	ErrInfraFailedPrecondition: http.StatusPreconditionFailed,
	ErrInfraConflict:           http.StatusConflict,
	ErrInfraNotModified:        http.StatusUnprocessableEntity,
	ErrInfraAborted:            http.StatusUnprocessableEntity,
	ErrInfraOutOfRange:         http.StatusUnprocessableEntity,
	ErrInfraNotImplemented:     http.StatusNotImplemented,
	ErrInfraUnavailable:        http.StatusConflict,
	ErrInfraDataLoss:           http.StatusUnprocessableEntity,
	ErrInfraUnauthorized:       http.StatusUnauthorized,
}

// errorWarnLevelMap defines the errors that are handled but unexpected to happen.
// Every error defined in this map will be notified at WARN level instead of ERROR.
var errorWarnLevelMap = map[error]bool{}

func RegisterStatusMapping(err error, statusCode int) {
	errorStatusMap[err] = statusCode
}
