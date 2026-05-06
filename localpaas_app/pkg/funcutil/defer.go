package funcutil

import (
	"errors"
	"fmt"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func EnsureNoPanic(currentErr *error) {
	if r := recover(); r != nil {
		panicErr := apperrors.NewPanic(fmt.Sprintf("%v", r))
		if currentErr != nil && *currentErr != nil {
			*currentErr = errors.Join(*currentErr, panicErr)
		} else if currentErr != nil {
			*currentErr = panicErr
		}
	}
}
