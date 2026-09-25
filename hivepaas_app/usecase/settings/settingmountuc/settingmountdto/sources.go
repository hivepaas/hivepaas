package settingmountdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type ListSettingMountSourcesReq struct {
}

func NewListSettingMountSourcesReq() *ListSettingMountSourcesReq {
	return &ListSettingMountSourcesReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ListSettingMountSourcesReq) Validate() hperrors.ValidationErrors {
	return nil
}

type ListSettingMountSourcesResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data []*SettingMountSource `json:"data"`
	// MayMountSensitive says whether the caller may mount a private key or a
	// password: the screen locks those parts, with the reason, when not.
	MayMountSensitive bool `json:"mayMountSensitive"`
}

// SettingMountSource is a type of setting an entry may mount from, and its parts.
type SettingMountSource struct {
	Type  base.SettingType    `json:"type"`
	Parts []*SettingMountPart `json:"parts"`
}

type SettingMountPart struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	Sensitive bool   `json:"sensitive"`
}
