package basedto

import (
	vld "github.com/tiendc/go-validator"
	vldbase "github.com/tiendc/go-validator/base"
)

func ValidateNumber[T vldbase.Number](v *T, required bool, min, max T, field string) (result []vld.Validator) {
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

func ValidateNumberIn[T vldbase.Number](v *T, required bool, allowedValues []T, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Must(v != nil).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if v != nil && (required || *v != 0) && len(allowedValues) > 0 {
		result = append(result,
			vld.NumIn(v, allowedValues...).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
			))
	}
	return result
}

func ValidateIntSliceRange[T vldbase.Number](s []T, min, max T, field string) (result []vld.Validator) {
	for _, value := range s {
		v := value
		result = append(result, vld.NumRange(&v, min, max).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_RANGE"),
		))
	}
	return result
}
