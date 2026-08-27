package basedto

import "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"

// TransformObjectSlice transforms a slice of objects
func TransformObjectSlice[T, U any, TS ~[]T](objects TS, singleTransformFn func(T) (U, error)) ([]U, error) {
	resp := make([]U, 0, len(objects))
	for _, obj := range objects {
		itemResp, err := singleTransformFn(obj)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp = append(resp, itemResp)
	}
	return resp, nil
}
