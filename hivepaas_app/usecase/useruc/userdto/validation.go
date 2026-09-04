package userdto

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

const (
	nameMinLen = 1
	nameMaxLen = 100

	notesMinLen = 1
	notesMaxLen = 10000
)

func validateUsername(username *string, required bool, field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateStr(username, required,
		nameMinLen, nameMaxLen, "username")...)

	// NOTE: username must not be a valid email address as it takes the address
	if username != nil {
		res = append(res, vld.Must(!govalidator.IsEmail(*username)).OnError(
			vld.SetField(field, nil),
			vld.SetCustomKey("ERR_ARGUMENT_INVALID"),
		))
	}
	return res
}

// validateProjectAccesses validates the per-project access grants of an invite or
// update request.
func validateProjectAccesses(projectAccesses []*ProjectAccessReq, field string) (res []vld.Validator) {
	for idx, projectAccess := range projectAccesses {
		itemField := fmt.Sprintf("%s[%d]", field, idx)
		res = append(res, basedto.ValidateObjectIDReq(&projectAccess.Project, true, itemField+".project")...)
		res = append(res, validateEnvAccesses(projectAccess.EnvAccesses, projectAccess.Project.ID,
			itemField+".envAccesses")...)
	}
	return res
}

// validateEnvAccesses validates a list of per-project-env access grants. User
// permissions are granted per project env only, so every ID must be a project
// env ID ("<projectID>:<envKey>"), never a bare project ID.
//
// Each env must also belong to projectID: the grants are stored from the env IDs
// alone, so an env from another project would silently be granted under the wrong
// project instead of being rejected.
func validateEnvAccesses(accesses basedto.ObjectAccessSliceReq, projectID, field string,
) (res []vld.Validator) {
	res = append(res, basedto.ValidateObjectAccessSliceReq(accesses, true, 0, field)...)
	for idx, access := range accesses {
		idField := fmt.Sprintf("%s[%d].id", field, idx)
		res = append(res, vld.Must(projecthelper.IsProjectEnvID(access.ID)).OnError(
			vld.SetField(idField, nil),
			vld.SetCustomKey("ERR_VLD_PROJECT_ENV_ID_INVALID"),
		))
		envProjectID, _ := projecthelper.ParseProjectEnvID(access.ID)
		res = append(res, vld.Must(envProjectID == projectID).OnError(
			vld.SetField(idField, nil),
			vld.SetCustomKey("ERR_VLD_PROJECT_ENV_ID_MISMATCHED"),
		))
	}
	return res
}

func validateUserPhoto(photo *UserPhotoReq, field string) (res []vld.Validator) {
	if photo == nil || photo.FileName == "" {
		return nil
	}
	fileExt := strings.ToLower(filepath.Ext(photo.FileName))
	return []vld.Validator{
		vld.Must(gofn.Contain(base.AllPhotoFileExts, fileExt)).OnError(
			vld.SetField(field+".fileName", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_EXT_UNSUPPORTED"),
		),
		vld.Must(len(photo.DataBytes) > 0 && base.IsValidPhotoContent(photo.DataBytes, fileExt)).OnError(
			vld.SetField(field+".dataBase64", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_INVALID"),
		),
		vld.When(len(photo.DataBytes) > 0).Then(
			vld.Must(len(photo.DataBytes) <= base.PhotoMaxSize).OnError(
				vld.SetField(field+".dataBase64", nil),
				vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_TOO_BIG"),
			),
		),
	}
}
