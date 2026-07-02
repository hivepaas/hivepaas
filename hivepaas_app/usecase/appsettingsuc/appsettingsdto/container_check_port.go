package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type CheckAppContainerPortReq struct {
	ProjectID string            `json:"-"`
	AppID     string            `json:"-"`
	Port      uint              `json:"port"`
	Timeout   timeutil.Duration `json:"timeout"`
}

func NewCheckAppContainerPortReq() *CheckAppContainerPortReq {
	return &CheckAppContainerPortReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CheckAppContainerPortReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CheckAppContainerPortResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *CheckAppContainerPortDataResp `json:"data"`
}

type CheckAppContainerPortDataResp struct {
	Open bool `json:"open"`
}
