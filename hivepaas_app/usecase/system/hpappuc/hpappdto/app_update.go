package hpappdto

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type UpdateHpAppReq struct {
	TargetVersion string `json:"targetVersion"`

	// SkipBackup updates without dumping the database first. See
	// entity.TaskSystemUpdateArgs for what that gives up.
	SkipBackup bool `json:"skipBackup"`
}

func NewUpdateHpAppReq() *UpdateHpAppReq {
	return &UpdateHpAppReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateHpAppReq) Validate() hperrors.ValidationErrors {
	return nil
}

type UpdateHpAppResp struct {
	Meta *basedto.Meta `json:"meta"`
}
