package basedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"
)

func ValidateDomain[T ~string](s *T, required bool, maxLen int, wildcardAllowed bool, field string) (
	result []vld.Validator) {
	if required {
		result = append(result, vld.Required(s).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if s != nil && *s != "" {
		result = append(result,
			vld.StrLen(s, 1, maxLen).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
			),
			vld.StrIsDNSName(s).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_DOMAIN_INVALID"),
			),
		)
		if !wildcardAllowed {
			result = append(result, vld.Must(!strings.Contains(string(*s), "*")).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_WILDCARD_UNALLOWED"),
			))
		}
	}
	return result
}

func ValidatePort[T int | uint | int32 | uint32 | int64 | uint64](v *T, required bool, min T,
	field string) []vld.Validator {
	return ValidateNumber(v, required, min, 65535, field) //nolint:mnd
}
