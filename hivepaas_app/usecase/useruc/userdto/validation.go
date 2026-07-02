package userdto

import (
	"path/filepath"
	"strings"

	"github.com/asaskevich/govalidator"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
)

const (
	nameMinLen = 1
	nameMaxLen = 100

	photoMaxSize = 300 * 1024 // 300KB

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
		vld.Must(len(photo.DataBytes) > 0).OnError(
			vld.SetField(field+".dataBase64", nil),
			vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_INVALID"),
		),
		vld.When(len(photo.DataBytes) > 0).Then(
			vld.Must(len(photo.DataBytes) <= photoMaxSize).OnError(
				vld.SetField(field+".dataBase64", nil),
				vld.SetCustomKey("ERR_VLD_USER_PHOTO_FILE_TOO_BIG"),
			),
		),
	}
}
