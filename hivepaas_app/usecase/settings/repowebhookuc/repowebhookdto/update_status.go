package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateRepoWebhookStatusReq struct {
	settings.UpdateSettingStatusReq
}

func NewUpdateRepoWebhookStatusReq() *UpdateRepoWebhookStatusReq {
	return &UpdateRepoWebhookStatusReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateRepoWebhookStatusReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateStrIn(req.Status, false,
		base.AllSettingSettableStatuses, "status")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRepoWebhookStatusResp struct {
	Meta *basedto.Meta `json:"meta"`
}
