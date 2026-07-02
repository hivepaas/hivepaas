package basedto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func ValidateTime(t *time.Time, required bool, from, to time.Time, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Required(t).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if t != nil && !t.IsZero() {
		dt := *t
		if !from.IsZero() {
			result = append(result, vld.TimeGTE(dt, from).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_MUST_GREATER_THAN"),
			))
		}
		if !to.IsZero() {
			result = append(result, vld.TimeLT(dt, to).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_MUST_LESS_THAN"),
			))
		}
	}
	return result
}

func ValidateDate(date *timeutil.Date, required bool, from, to timeutil.Date, field string) (result []vld.Validator) {
	var t *time.Time
	if date != nil {
		t = new(date.ToTime())
	}
	return ValidateTime(t, required, from.ToTime(), to.ToTime(), field)
}
