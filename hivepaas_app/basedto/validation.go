package basedto

import (
	"math"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"
)

func ValidateRequiredField(field any, fieldName string) (res []vld.Validator) {
	res = append(res, vld.Required(field).OnError(
		vld.SetField(fieldName, nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
	))
	return res
}

func ValidateSlice[T comparable, S ~[]T](s S, unique bool, minLen int, allowedValues []T, field string) (
	result []vld.Validator) {
	return ValidateSliceEx(s, unique, minLen, math.MaxInt, allowedValues, field)
}

func ValidateSliceEx[T comparable, S ~[]T](s S, unique bool, minLen, maxLen int, allowedValues []T, field string) (
	result []vld.Validator) {
	if unique {
		result = append(result, vld.SliceUnique(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
		))
	}

	// Check minimum length
	if minLen > 0 && len(s) < minLen {
		result = append(result, vld.SliceLen(s, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}

	// Check maximum length separately with specific error message
	if maxLen < math.MaxInt && len(s) > maxLen {
		result = append(result, vld.Must(len(s) <= maxLen).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_TOO_MANY"),
			vld.SetParam("Max", maxLen),
			vld.SetParam("Actual", len(s)),
		))
	}

	if len(allowedValues) > 0 {
		result = append(result,
			vld.SliceElemIn(s, allowedValues...).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
			))
	}
	return result
}

func ValidateObjectSliceBy[T, A comparable, S ~[]T](s S, required, unique bool, allowedValues []A,
	mapBy func(T) A, field string) (result []vld.Validator) {
	if required && len(s) == 0 {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}

	values := gofn.MapSlice(s, func(s T) A { return mapBy(s) })
	if unique {
		result = append(result, vld.SliceUnique(values).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
		))
	}
	if len(s) > 0 && len(allowedValues) > 0 {
		result = append(result,
			vld.SliceElemIn(values, allowedValues...).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
			))
	}
	return result
}

func ValidateCond(mustCond bool, field string) (result []vld.Validator) {
	result = append(result,
		vld.Must(mustCond).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_INVALID"),
		),
	)
	return result
}

func ValidateMutualExclusiveFields(mustCond bool, fields ...string) (result []vld.Validator) {
	for _, field := range fields {
		result = append(result,
			vld.Must(mustCond).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_MUTUALLY_EXCLUSIVE_FIELDS_INVALID"),
			),
		)
	}
	return result
}
