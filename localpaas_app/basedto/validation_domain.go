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
