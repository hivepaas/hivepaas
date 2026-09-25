package settingmountdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	entryKeyMaxLen = 20
	pathMaxLen     = 512
	maxFiles       = 20
)

// SettingMountBaseReq is an entry as the screen writes it. What the source's
// type allows is checked by the use case, which reads the source.
type SettingMountBaseReq struct {
	// Name is the entry's key.
	Name   string                 `json:"name"`
	Source basedto.ObjectIDReq    `json:"source"`
	Files  []*SettingMountFileReq `json:"files"`
}

type SettingMountFileReq struct {
	Part string            `json:"part"`
	Path string            `json:"path"`
	UID  string            `json:"uid"`
	GID  string            `json:"gid"`
	Mode fileutil.FileMode `json:"mode"`
}

func (req *SettingMountBaseReq) ToEntity() *entity.AppSettingMount {
	mount := &entity.AppSettingMount{Source: entity.ObjectID{ID: req.Source.ID}}
	for _, f := range req.Files {
		if f == nil {
			continue
		}
		mount.Files = append(mount.Files, &entity.AppSettingMountFile{
			Part: f.Part, Path: f.Path, UID: f.UID, GID: f.GID, Mode: f.Mode,
		})
	}
	return mount
}

func (req *SettingMountBaseReq) validate() (res []vld.Validator) {
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, entryKeyMaxLen, "name")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.Source, true, "source")...)
	res = append(res, vld.SliceLen(req.Files, 1, maxFiles).OnError(
		vld.SetField("files", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
	))
	for _, f := range req.Files {
		if f == nil {
			continue
		}
		res = append(res, basedto.ValidateStr(&f.Path, true, 1, pathMaxLen, "files.path")...)
	}
	return res
}

type CreateSettingMountReq struct {
	settings.CreateSettingReq
	*SettingMountBaseReq
}

func NewCreateSettingMountReq() *CreateSettingMountReq {
	return &CreateSettingMountReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	if req.SettingMountBaseReq == nil {
		req.SettingMountBaseReq = &SettingMountBaseReq{}
	}
	validators = append(validators, req.validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateSettingMountResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
