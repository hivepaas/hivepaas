package hperrors

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

// warnLevelErrors are errors that are handled but unexpected to happen: they answer
// the caller with a 4xx - it did ask for something that cannot be done - yet still
// leave the system in a state a person has to repair. They are reported at WARN
// instead of the INFO their status would otherwise earn them, which is what decides
// whether they are recorded at all.
//
// Matching is errors.Is against the entry, so an entry can be one specific error or a
// whole class. Prefer the specific error: naming a base error such as
// ErrPreconditionFailed raises every business rule built on it, and those are the
// ordinary refusals this list exists to be distinguished from.
var warnLevelErrors = []error{
	// The repository was re-encrypted, storing the new password failed, and putting
	// the old one back failed too. Nothing recovers this except a person.
	ErrBackupRepoPasswordOutOfSync,
	// Data written by a newer version than the one running: a rollback or a deploy
	// went sideways, and the next write may lose fields it does not understand.
	ErrDataVerNewerThanSystemVer,
}

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
	ErrGRPCRequestFailed        = NewErr(ErrActionFailed, "ERR_GRPC_REQUEST_FAILED")
	ErrUnavailable              = NewErr(ErrPreconditionFailed, "ERR_UNAVAILABLE")
	ErrArgumentInvalid          = NewErr(ErrBadRequest, "ERR_ARGUMENT_INVALID")
	ErrValueInvalid             = NewErr(ErrPreconditionFailed, "ERR_VALUE_INVALID")
	ErrTokenInvalid             = NewErr(ErrValueInvalid, "ERR_TOKEN_INVALID")
	ErrMismatch                 = NewErr(ErrPreconditionFailed, "ERR_MISMATCH")
	ErrUpdateVerMismatched      = NewErr(ErrMismatch, "ERR_UPDATE_VER_MISMATCHED")
	ErrValidation               = NewErr(ErrBadRequest, "ERR_VALIDATION")
)
