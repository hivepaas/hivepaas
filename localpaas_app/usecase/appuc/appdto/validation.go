package appdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	appNameMinLen = 1
	appNameMaxLen = 100

	appEnvMinLen = 1
	appEnvMaxLen = 50

	appTagMinLen = 0
	appTagMaxLen = 50

	appNoteMinLen = 1
	appNoteMaxLen = 10000
)

func validateAppName(name *string, field string) []vld.Validator {
	return basedto.ValidateStr(name, true, appNameMinLen, appNameMaxLen, field)
	// TODO: need validation for valid characters
}

func validateAppEnv(env *string, field string) []vld.Validator {
	return basedto.ValidateStr(env, false, appEnvMinLen, appEnvMaxLen, field)
}

func validateAppTags(tags []string, field string) []vld.Validator {
	return basedto.ValidateSliceEx(tags, true, appTagMinLen, appTagMaxLen, nil, field)
}

func validateAppNote(note *string, field string) []vld.Validator {
	return basedto.ValidateStr(note, false, appNoteMinLen, appNoteMaxLen, field)
}
