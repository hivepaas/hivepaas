package basedto

import (
	"encoding/base64"

	vld "github.com/tiendc/go-validator"
)

func ValidateStr[T ~string](s *T, required bool, minLen, maxLen int, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		result = append(result, vld.StrLen(s, minLen, maxLen).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
		))
	}
	return result
}

func ValidateStrIn[T ~string](s *T, required bool, allowedValues []T, field string) (
	result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && len(*s) > 0 && len(allowedValues) > 0 {
		result = append(result,
			vld.StrIn(s, allowedValues...).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
			))
	}
	return result
}

func ValidateStrNotIn[T ~string](s *T, required bool, minLen, maxLen int, unallowedValues []T, field string) (
	result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		result = append(result, vld.StrLen(s, minLen, maxLen).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
		))

		if len(unallowedValues) > 0 {
			result = append(result,
				vld.StrNotIn(s, unallowedValues...).OnError(
					vld.SetField(field, nil),
					vld.SetCustomKey("ERR_VLD_VALUE_UNALLOWED"),
				))
		}
	}

	return result
}

func ValidateStrBase64[T ~string](s *T, required bool, minLen, maxLen int, field string) (result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		sVal, err := base64.StdEncoding.DecodeString(string(*s))
		if err == nil {
			result = append(result, vld.Must(len(sVal) >= minLen && len(sVal) <= maxLen).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
				vld.SetParam("Min", minLen),
				vld.SetParam("Max", maxLen),
			))
		} else {
			result = append(result, vld.Must(false).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_VALUE_IS_NOT_OF_TYPE"),
				vld.SetParam("Type", "base64"),
			))
		}
	}
	return result
}
