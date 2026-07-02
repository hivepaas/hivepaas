package basedto

import (
	"fmt"
	"math"

	vld "github.com/tiendc/go-validator"
)

func ValidateEmail(email *string, required bool, field string) (res []vld.Validator) {
	if required {
		res = append(res, vld.Required(email).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_EMAIL_ADDR_REQUIRED"), // use default error key
		))
	}
	if email != nil && *email != "" {
		res = append(res, vld.StrIsEmail(email).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_EMAIL_ADDR_INVALID"), // use default error key
		))
	}
	return res
}

func ValidateEmailSlice(emails []string, unique bool, minLen int, field string) (result []vld.Validator) {
	if unique {
		result = append(result, vld.SliceUnique(emails).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_EMAIL_ADDRS_NON_UNIQUE"), // use default error key
		))
	}
	if minLen > 0 {
		result = append(result, vld.SliceLen(emails, minLen, math.MaxInt).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_EMAIL_ADDRS_REQUIRED"), // use default error key
		))
	}
	result = append(result,
		vld.Slice(emails).ForEach(func(element string, index int, elemValidator vld.ItemValidator) {
			elemValidator.Validate(
				vld.StrIsEmail(&element).OnError(
					vld.SetField(fmt.Sprintf("%s[%d]", field, index), nil),
					vld.SetCustomKey("ERR_VLD_EMAIL_ADDR_INVALID"), // use default error key
				),
			)
		}),
	)
	return result
}
