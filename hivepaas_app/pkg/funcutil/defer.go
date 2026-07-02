package funcutil

import (
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func EnsureNoPanic(currentErr *error) {
	if r := recover(); r != nil {
		panicErr := apperrors.NewPanic(r)
		if currentErr != nil && *currentErr != nil {
			*currentErr = errors.Join(*currentErr, panicErr)
		} else if currentErr != nil {
			*currentErr = panicErr
		}
	}
}
