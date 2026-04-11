package projectsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

const (
	minProjectNameLen = 1
	maxProjectNameLen = 100

	minProjectTagLen = 0
	maxProjectTagLen = 100

	minProjectNoteLen = 10
	maxProjectNoteLen = 10000
)

func validateProjectName(name *string, field string) []vld.Validator {
	return basedto.ValidateStr(name, true, minProjectNameLen, maxProjectNameLen, field)
	// TODO: need validation for valid characters
}

func validateProjectTags(tags []string, field string) []vld.Validator {
	return basedto.ValidateSliceEx(tags, true, minProjectTagLen, maxProjectTagLen, nil, field)
}

func validateProjectNote(note *string, field string) []vld.Validator {
	return basedto.ValidateStr(note, false, minProjectNoteLen, maxProjectNoteLen, field)
}

func validateProjectOwner(id *basedto.ObjectIDReq, field string) []vld.Validator {
	return basedto.ValidateObjectIDReq(id, false, field)
}
