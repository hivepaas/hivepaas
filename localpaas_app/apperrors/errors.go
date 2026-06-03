package apperrors

import (
	"errors"
	"net/http"

	"github.com/localpaas/localpaas/localpaas_app/pkg/translation"
)

// Wrap wraps an error with storing stack trace
func Wrap(err error) error {
	return New(err)
}

type ErrLevel uint8

const (
	ErrLevelInfo  ErrLevel = iota + 1
	ErrLevelWarn  ErrLevel = iota + 1
	ErrLevelError ErrLevel = iota + 1
)

// ParseError parse the given error and return a list of ErrorInfo.
// If the given error is a single one, the returned slice will contain only one item.
func ParseError(err error, lang translation.Lang) (*ErrorInfo, ErrLevel) {
	// err is ValidationErrors
	if validationErrs, ok := errors.AsType[ValidationErrors](err); ok {
		return validationErrs.Build(lang), ErrLevelInfo
	}

	// `New` will automatically create AppError if the input is not AppError
	appErr := New(err)
	errorInfo := appErr.Build(lang)
	if errorInfo.Status == http.StatusInternalServerError {
		return errorInfo, ErrLevelError
	}
	if errorWarnLevelMap[appErr.UnwrapTilRoot()] {
		return errorInfo, ErrLevelWarn
	}
	// User error, not the logic and not unexpected, reports at INFO level
	return errorInfo, ErrLevelInfo
}

// ParseErrorDetail parses to get detail from the given error
func ParseErrorDetail(err error, lang translation.Lang) (detail string) {
	errInfo, _ := ParseError(err, lang)
	if errInfo != nil {
		detail = errInfo.Detail
	}
	return detail
}

// NewPanic return AppError for error Panic
func NewPanic(err any) AppError {
	return New(ErrPanic).WithNTParam("Error", err)
}

// NewNotFound return AppError for error NotFound
func NewNotFound(name any) AppError {
	return New(ErrNotFound).WithParam("Name", name)
}
func NewNotFoundNT(name any) AppError { // NT: non translation param
	return New(ErrNotFound).WithNTParam("Name", name)
}

// NewAlreadyExist return AppError for error AlreadyExist
func NewAlreadyExist(name any) AppError {
	return New(ErrAlreadyExist).WithParam("Name", name)
}
func NewAlreadyExistNT(name any) AppError { // NT: non translation param
	return New(ErrAlreadyExist).WithNTParam("Name", name)
}

// NewConflict return AppError for error Conflict
func NewConflict(name any) AppError {
	return New(ErrConflict).WithParam("Name", name)
}
func NewConflictNT(name any) AppError { // NT: non translation param
	return New(ErrConflict).WithNTParam("Name", name)
}

// NewParamInvalid return AppError for error ParamInvalid
func NewParamInvalid(name any) AppError {
	return New(ErrParamInvalid).WithParam("Name", name)
}
func NewParamInvalidNT(name any) AppError { // NT: non translation param
	return New(ErrParamInvalid).WithNTParam("Name", name)
}

// NewUnavailable return AppError for error Unavailable
func NewUnavailable(name any) AppError {
	return New(ErrUnavailable).WithParam("Name", name)
}
func NewUnavailableNT(name any) AppError { // NT: non translation param
	return New(ErrUnavailable).WithNTParam("Name", name)
}

// NewForbidden return AppError for error Forbidden
func NewForbidden(name any) AppError {
	return New(ErrForbidden).WithParam("Name", name)
}
func NewForbiddenNT(name any) AppError { // NT: non translation param
	return New(ErrForbidden).WithNTParam("Name", name)
}

// NewNonEditable return AppError for error NonEditable
func NewNonEditable(name any) AppError {
	return New(ErrNonEditable).WithParam("Name", name)
}
func NewNonEditableNT(name any) AppError { // NT: non translation param
	return New(ErrNonEditable).WithNTParam("Name", name)
}

// NewNonDeletable return AppError for error NonDeletable
func NewNonDeletable(name any) AppError {
	return New(ErrNonDeletable).WithParam("Name", name)
}
func NewNonDeletableNT(name any) AppError { // NT: non translation param
	return New(ErrNonDeletable).WithNTParam("Name", name)
}

// NewInUse return AppError for error ResourceInUse
func NewInUse(name any) AppError {
	return New(ErrResourceInUse).WithParam("Name", name)
}
func NewInUseNT(name any) AppError { // NT: non translation param
	return New(ErrResourceInUse).WithNTParam("Name", name)
}

// NewInactive return AppError for error ResourceInactive
func NewInactive(name any) AppError {
	return New(ErrResourceInactive).WithParam("Name", name)
}
func NewInactiveNT(name any) AppError { // NT: non translation param
	return New(ErrResourceInactive).WithNTParam("Name", name)
}

// NewMissing return AppError for error ResourceMissing
func NewMissing(name any) AppError {
	return New(ErrResourceMissing).WithParam("Name", name)
}
func NewMissingNT(name any) AppError { // NT: non translation param
	return New(ErrResourceMissing).WithNTParam("Name", name)
}

// NewTypeInvalid return AppError for error TypeInvalid
func NewTypeInvalid() AppError {
	return New(ErrTypeInvalid)
}
func NewTypeInvalidNT() AppError { // NT: non translation param
	return New(ErrTypeInvalid)
}

// NewValueInvalid return AppError for error ValueInvalid
func NewValueInvalid() AppError {
	return New(ErrValueInvalid)
}
func NewValueInvalidNT() AppError { // NT: non translation param
	return New(ErrValueInvalid)
}

// NewUnsupported return AppError for error Unsupported
func NewUnsupported(name any) AppError {
	return New(ErrUnsupported).WithParam("Name", name)
}
func NewUnsupportedNT(name any) AppError { // NT: non translation param
	return New(ErrUnsupported).WithNTParam("Name", name)
}

// NewNotImplemented return AppError for error NotImplemented
func NewNotImplemented() AppError {
	return New(ErrNotImplemented)
}
func NewNotImplementedNT() AppError { // NT: non translation param
	return New(ErrNotImplemented)
}
