package projectdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	projectNameMinLen = 1
	projectNameMaxLen = 100

	projectEnvMinLen = 1
	projectEnvMaxLen = 50

	projectTagMinLen = 0
	projectTagMaxLen = 50

	projectNoteMinLen = 1
	projectNoteMaxLen = 10000
)

func validateProjectName(name *string, field string) []vld.Validator {
	return basedto.ValidateStr(name, true, projectNameMinLen, projectNameMaxLen, field)
	// TODO: need validation for valid characters
}

func validateProjectEnvs(envs []*ProjectEnvReq, field string) (res []vld.Validator) {
	for i, env := range envs {
		res = append(res, basedto.ValidateStr(&env.Name, true, projectEnvMinLen, projectEnvMaxLen,
			field+fmt.Sprintf("[%v].name", i))...)
	}
	res = append(res, vld.SliceUniqueBy(envs, func(env *ProjectEnvReq) string { return env.Name }))
	return res
}

func validateProjectTags(tags []string, field string) []vld.Validator {
	return basedto.ValidateSliceEx(tags, true, projectTagMinLen, projectTagMaxLen, nil, field)
}

func validateProjectNote(note *string, field string) []vld.Validator {
	return basedto.ValidateStr(note, false, projectNoteMinLen, projectNoteMaxLen, field)
}

func validateProjectOwner(id *basedto.ObjectIDReq, field string) []vld.Validator {
	return basedto.ValidateObjectIDReq(id, false, field)
}
