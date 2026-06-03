package funcutil

import (
	"errors"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func EnsureNoPanic(currentErr *error) {
	if r := recover(); r != nil {
		panicErr := apperrors.NewPanic(apperrors.Fmt("%v", r))
		if currentErr != nil && *currentErr != nil {
			*currentErr = errors.Join(*currentErr, panicErr)
		} else if currentErr != nil {
			*currentErr = panicErr
		}
	}
}
