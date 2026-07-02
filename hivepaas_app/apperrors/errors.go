package apperrors

import (
	"errors"
	"net/http"

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
	baseErr := getBaseError(appErr)
	if baseErr != nil && errorWarnLevelMap[baseErr] {
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

// NewInternal return AppError for error Internal
func NewInternal() AppError {
	return New(ErrInternal)
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

// NewArgumentInvalid return AppError for error ErrArgumentInvalid
func NewArgumentInvalid(name any) AppError {
	return New(ErrArgumentInvalid).WithParam("Name", name)
}
func NewArgumentInvalidNT(name any) AppError { // NT: non translation param
	return New(ErrArgumentInvalid).WithNTParam("Name", name)
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
	return New(ErrInUse).WithParam("Name", name)
}
func NewInUseNT(name any) AppError { // NT: non translation param
	return New(ErrInUse).WithNTParam("Name", name)
}

// NewInactive return AppError for error ResourceInactive
func NewInactive(name any) AppError {
	return New(ErrInactive).WithParam("Name", name)
}
func NewInactiveNT(name any) AppError { // NT: non translation param
	return New(ErrInactive).WithNTParam("Name", name)
}

// NewMissing return AppError for error ResourceMissing
func NewMissing(name any) AppError {
	return New(ErrMissing).WithParam("Name", name)
}
func NewMissingNT(name any) AppError { // NT: non translation param
	return New(ErrMissing).WithNTParam("Name", name)
}

// NewMismatch return AppError for error Mismatch
func NewMismatch(left, right any) AppError {
	return New(ErrMismatch).WithParam("Left", left).WithParam("Right", right)
}
func NewMismatchNT(left, right any) AppError { // NT: non translation param
	return New(ErrMismatch).WithNTParam("Left", left).WithNTParam("Right", right)
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

// ToGRPCError converts any error (including AppError) to a gRPC status error.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// If it is already a gRPC status error, return as is
	if _, ok := status.FromError(err); ok {
		return err
	}

	if appErr, ok := errors.AsType[AppError](err); ok {
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
