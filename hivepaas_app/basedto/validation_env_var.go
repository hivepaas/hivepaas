package basedto

import (
	"fmt"
	"regexp"

	vld "github.com/tiendc/go-validator"
)

const (
	envNameMaxLen = 100
)

var envNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func ValidateEnvVarsReq(envVars []*EnvVarReq, field string) (res []vld.Validator) {
	res = append(res, vld.SliceUniqueBy(envVars, func(t *EnvVarReq) string {
		return t.Key
	}).OnError(
		vld.SetField(field, nil),
		vld.SetCustomKey("ERR_VLD_VALUES_NON_UNIQUE"),
	))
	for i, env := range envVars {
		res = append(res, ValidateEnvName(&env.Key, true, fmt.Sprintf("%s[%d].key", field, i))...)
	}
	return res
}

func ValidateEnvName(v *string, required bool, field string) (res []vld.Validator) {
	if required {
		res = append(res, vld.Required(v).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
		))
	}
	if v != nil && (required || *v != "") {
		res = append(res,
			vld.StrLen(v, 1, envNameMaxLen).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_FIELD_LENGTH_INVALID"),
			),
			vld.StrByteMatch(v, envNameRegex).OnError(
				vld.SetField(field, nil),
				vld.SetCustomKey("ERR_VLD_NAME_INVALID"),
			),
		)
	}
	return res
}
