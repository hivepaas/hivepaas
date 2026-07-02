package apperrors

import (
	"errors"

	"github.com/containerd/errdefs"
)

var (
	IsInfraUnknown            = errdefs.IsUnknown
	IsInfraInvalidArgument    = errdefs.IsInvalidArgument
	IsInfraNotFound           = errdefs.IsNotFound
	IsInfraAlreadyExists      = errdefs.IsAlreadyExists
	IsInfraPermissionDenied   = errdefs.IsPermissionDenied
	IsInfraResourceExhausted  = errdefs.IsResourceExhausted
	IsInfraFailedPrecondition = errdefs.IsFailedPrecondition
	IsInfraConflict           = errdefs.IsConflict
	IsInfraNotModified        = errdefs.IsNotModified
	IsInfraAborted            = errdefs.IsAborted
	IsInfraOutOfRange         = errdefs.IsOutOfRange
	IsInfraNotImplemented     = errdefs.IsNotImplemented
	IsInfraInternal           = errdefs.IsInternal
	IsInfraUnavailable        = errdefs.IsUnavailable
	IsInfraDataLoss           = errdefs.IsDataLoss
	IsInfraUnauthorized       = errdefs.IsUnauthorized
)

func NewInfra(err error) AppError {
	if err == nil {
		return nil
	}
	infraErr := ErrInfra
	switch {
	case IsInfraUnknown(err):
		infraErr = ErrInfraUnknown
	case IsInfraInvalidArgument(err):
		infraErr = ErrInfraInvalidArgument
	case IsInfraNotFound(err):
		infraErr = ErrInfraNotFound
	case IsInfraAlreadyExists(err):
		infraErr = ErrInfraAlreadyExists
	case IsInfraPermissionDenied(err):
		infraErr = ErrInfraPermissionDenied
	case IsInfraResourceExhausted(err):
		infraErr = ErrInfraResourceExhausted
	case IsInfraFailedPrecondition(err):
		infraErr = ErrInfraFailedPrecondition
	case IsInfraConflict(err):
		infraErr = ErrInfraConflict
	case IsInfraNotModified(err):
		infraErr = ErrInfraNotModified
	case IsInfraAborted(err):
		infraErr = ErrInfraAborted
	case IsInfraOutOfRange(err):
		infraErr = ErrInfraOutOfRange
	case IsInfraNotImplemented(err):
		infraErr = ErrInfraNotImplemented
	case IsInfraInternal(err):
		infraErr = ErrInfraInternal
	case IsInfraUnavailable(err):
		infraErr = ErrInfraUnavailable
	case IsInfraDataLoss(err):
		infraErr = ErrInfraDataLoss
	case IsInfraUnauthorized(err):
		infraErr = ErrInfraUnauthorized
	}
	return New(errors.Join(infraErr, err)).WithNTParam("Error", err.Error())
}
