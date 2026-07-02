package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type UpdateRepoWebhookReq struct {
	settings.UpdateSettingReq
	*RepoWebhookBaseReq
}

func NewUpdateRepoWebhookReq() *UpdateRepoWebhookReq {
	return &UpdateRepoWebhookReq{}
}

func (req *UpdateRepoWebhookReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateRepoWebhookReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRepoWebhookResp struct {
	Meta *basedto.Meta `json:"meta"`
}
