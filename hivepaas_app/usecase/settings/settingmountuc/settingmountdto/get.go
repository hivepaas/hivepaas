package settingmountdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetSettingMountReq struct {
	settings.GetSettingReq
}

func NewGetSettingMountReq() *GetSettingMountReq {
	return &GetSettingMountReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetSettingMountReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetSettingMountResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *SettingMountResp `json:"data"`
}

type SettingMountResp struct {
	*settings.BaseSettingResp
	Source *settings.BaseSettingResp `json:"source"`
	Files  []*SettingMountFileResp   `json:"files"`
	// State is what of the entry is mounted; null when it could not be read.
	State *settingmountservice.EntryState `json:"state"`
}

type SettingMountFileResp struct {
	Part      string            `json:"part"`
	Path      string            `json:"path"`
	UID       string            `json:"uid"`
	GID       string            `json:"gid"`
	Mode      fileutil.FileMode `json:"mode"`
	Sensitive bool              `json:"sensitive"`
}

// TransformSettingMount renders an entry with its source and its state.
func TransformSettingMount(
	setting *entity.Setting, refObjects *entity.RefObjects, state *settingmountservice.EntryState,
) (*SettingMountResp, error) {
	mount, err := setting.AsAppSettingMount()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &SettingMountResp{State: state, Files: make([]*SettingMountFileResp, 0, len(mount.Files))}
	if resp.BaseSettingResp, err = settings.TransformSettingBase(setting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if refObjects != nil && refObjects.RefSettings[mount.Source.ID] != nil {
		resp.Source, _ = settings.TransformSettingBase(refObjects.RefSettings[mount.Source.ID])
	}
	if resp.Source == nil {
		resp.Source = settings.NewMissingSetting(mount.Source.ID, "")
	}
	for _, f := range mount.Files {
		if f == nil {
			continue
		}
		resp.Files = append(resp.Files, &SettingMountFileResp{Part: f.Part, Path: f.Path, UID: f.UID,
			GID: f.GID, Mode: f.Mode, Sensitive: settingmountservice.SensitivePart(f.Part)})
	}
	return resp, nil
}
