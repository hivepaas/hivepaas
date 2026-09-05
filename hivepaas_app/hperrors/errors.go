package hperrors

import (
	"errors"
	"net/http"

	goerrors "github.com/go-errors/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

type ErrLevel uint8

const (
	ErrLevelInfo  ErrLevel = iota + 1
	ErrLevelWarn  ErrLevel = iota + 1
	ErrLevelError ErrLevel = iota + 1
)

// ShouldRecord reports whether an error at this level is worth storing for someone
// to look at later.
//
// Everything below WARN is the caller being told no - a failed validation, an expired
// token, a permission denied, a business rule refusing an action. Nobody can act on
// those, and anyone who can reach the API can produce them at will, so recording them
// turns the store into a log of other people's mistakes and hands an attacker a way
// to grow the database from outside.
func (l ErrLevel) ShouldRecord() bool {
	return l >= ErrLevelWarn
}

// ParseError parse the given error and return a list of ErrorInfo.
// If the given error is a single one, the returned slice will contain only one item.
func ParseError(err error, lang translation.Lang) (*ErrorInfo, ErrLevel) {
	// err is ValidationErrors
	if validationErrs, ok := errors.AsType[ValidationErrors](err); ok {
		return validationErrs.Build(lang), ErrLevelInfo
	}

	// `New` will automatically create AppError if the input is not AppError
	appErr := Wrap(err)
	errorInfo := appErr.Build(lang)
	// Any 5xx is ours, not the caller's: 500 for a bug, 503 when a dependency is
	// down, 501 for a path that was never finished. Matching only 500 would file an
	// outage as an ordinary user error.
	if errorInfo.Status >= http.StatusInternalServerError {
		return errorInfo, ErrLevelError
	}
	if isWarnLevelError(appErr) {
		return errorInfo, ErrLevelWarn
	}
	// User error, not the logic and not unexpected, reports at INFO level
	return errorInfo, ErrLevelInfo
}

// isWarnLevelError reports whether the error is one of the errors that are unexpected
// despite the status they carry.
func isWarnLevelError(err error) bool {
	for _, warnErr := range warnLevelErrors {
		if errors.Is(err, warnErr) {
			return true
		}
	}
	return false
}

// GetErrorDetail parses to get detail from the given error
func GetErrorDetail(err error, lang translation.Lang) string {
	if err == nil {
		return ""
	}
	if lang == "" {
		lang = translation.GetDefaultLang()
	}
	errInfo, _ := ParseError(err, lang)
	if errInfo == nil {
		return err.Error()
	}
	return errInfo.Code + "\n" + errInfo.Detail
}

func GetErrorStackTrace(err error) string {
	if errWithStack, ok := errors.AsType[*goerrors.Error](err); ok {
		return errWithStack.ErrorStack()
	}
	return ""
}

// NewInternal return HPError for error Internal
func NewInternal() HPError {
	return Wrap(ErrInternal)
}

// NewPanic return HPError for error Panic
func NewPanic(err any) HPError {
	return Wrap(ErrPanic).WithNTParam("Error", err)
}

// NewNotFound return HPError for error NotFound
func NewNotFound(name any) HPError {
	return Wrap(ErrNotFound).WithParam("Name", name)
}
func NewNotFoundNT(name any) HPError { // NT: non translation param
	return Wrap(ErrNotFound).WithNTParam("Name", name)
}

// NewAlreadyExist return HPError for error AlreadyExist
func NewAlreadyExist(name any) HPError {
	return Wrap(ErrAlreadyExist).WithParam("Name", name)
}
func NewAlreadyExistNT(name any) HPError { // NT: non translation param
	return Wrap(ErrAlreadyExist).WithNTParam("Name", name)
}

// NewConflict return HPError for error Conflict
func NewConflict(name any) HPError {
	return Wrap(ErrConflict).WithParam("Name", name)
}
func NewConflictNT(name any) HPError { // NT: non translation param
	return Wrap(ErrConflict).WithNTParam("Name", name)
}

// NewArgumentInvalid return HPError for error ErrArgumentInvalid
func NewArgumentInvalid(name any) HPError {
	return Wrap(ErrArgumentInvalid).WithParam("Name", name)
}
func NewArgumentInvalidNT(name any) HPError { // NT: non translation param
	return Wrap(ErrArgumentInvalid).WithNTParam("Name", name)
}

// NewUnavailable return HPError for error Unavailable
func NewUnavailable(name any) HPError {
	return Wrap(ErrUnavailable).WithParam("Name", name)
}
func NewUnavailableNT(name any) HPError { // NT: non translation param
	return Wrap(ErrUnavailable).WithNTParam("Name", name)
}

// NewForbidden return HPError for error Forbidden
func NewForbidden(name any) HPError {
	return Wrap(ErrForbidden).WithParam("Name", name)
}
func NewForbiddenNT(name any) HPError { // NT: non translation param
	return Wrap(ErrForbidden).WithNTParam("Name", name)
}

// NewNonEditable return HPError for error NonEditable
func NewNonEditable(name any) HPError {
	return Wrap(ErrNonEditable).WithParam("Name", name)
}
func NewNonEditableNT(name any) HPError { // NT: non translation param
	return Wrap(ErrNonEditable).WithNTParam("Name", name)
}

// NewNonDeletable return HPError for error NonDeletable
func NewNonDeletable(name any) HPError {
	return Wrap(ErrNonDeletable).WithParam("Name", name)
}
func NewNonDeletableNT(name any) HPError { // NT: non translation param
	return Wrap(ErrNonDeletable).WithNTParam("Name", name)
}

// NewInUse return HPError for error ResourceInUse
func NewInUse(name any) HPError {
	return Wrap(ErrInUse).WithParam("Name", name)
}
func NewInUseNT(name any) HPError { // NT: non translation param
	return Wrap(ErrInUse).WithNTParam("Name", name)
}

// NewInactive return HPError for error ResourceInactive
func NewInactive(name any) HPError {
	return Wrap(ErrInactive).WithParam("Name", name)
}
func NewInactiveNT(name any) HPError { // NT: non translation param
	return Wrap(ErrInactive).WithNTParam("Name", name)
}

// NewMissing return HPError for error ResourceMissing
func NewMissing(name any) HPError {
	return Wrap(ErrMissing).WithParam("Name", name)
}
func NewMissingNT(name any) HPError { // NT: non translation param
	return Wrap(ErrMissing).WithNTParam("Name", name)
}

// NewMismatch return HPError for error Mismatch
func NewMismatch(left, right any) HPError {
	return Wrap(ErrMismatch).WithParam("Left", left).WithParam("Right", right)
}
func NewMismatchNT(left, right any) HPError { // NT: non translation param
	return Wrap(ErrMismatch).WithNTParam("Left", left).WithNTParam("Right", right)
}

// NewUnsupported return HPError for error Unsupported
func NewUnsupported(name any) HPError {
	return Wrap(ErrUnsupported).WithParam("Name", name)
}
func NewUnsupportedNT(name any) HPError { // NT: non translation param
	return Wrap(ErrUnsupported).WithNTParam("Name", name)
}

// NewNotImplemented return HPError for error NotImplemented
func NewNotImplemented() HPError {
	return Wrap(ErrNotImplemented)
}
func NewNotImplementedNT() HPError { // NT: non translation param
	return Wrap(ErrNotImplemented)
}

// ToGRPCError converts any error (including HPError) to a gRPC status error.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// If it is already a gRPC status error, return as is
	if _, ok := status.FromError(err); ok {
		return err
	}

	if appErr, ok := errors.AsType[HPError](err); ok {
		grpcCode := grpcErrorStatusMap[getBaseError(appErr)]

		// Translate the error message using Default English Language
		detail, _ := appErr.Message(translation.LangEn)
		if detail == "" {
			detail = appErr.Error()
		}

		return status.Error(grpcCode, detail) //nolint:wrapcheck
	}

	// Fallback to internal error code
	return status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
}
