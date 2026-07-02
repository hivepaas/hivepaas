package basedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func ValidateDuration[T time.Duration | timeutil.Duration](v *T, required bool, min, max T, field string) (
	result []vld.Validator) {
	if required {
		result = append(result, vld.Must(v != nil).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if v != nil && (required || *v != 0) {
		result = append(result, vld.NumRange(v, min, max).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_RANGE"),
		))
	}
	return result
}
