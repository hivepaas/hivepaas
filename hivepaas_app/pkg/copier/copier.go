package copier

import (
	"github.com/tiendc/go-deepcopy"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func Copy(dst, src any) error {
	return deepcopy.Copy(dst, src) //nolint:wrapcheck
}

func CopyAs[T any](entity T) (copied T, err error) {
	if err = deepcopy.Copy(&copied, &entity); err != nil {
		return copied, apperrors.New(err)
	}
	return copied, nil
}
