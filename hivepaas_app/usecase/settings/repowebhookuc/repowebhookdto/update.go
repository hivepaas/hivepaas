package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
func (req *UpdateRepoWebhookReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRepoWebhookResp struct {
	Meta *basedto.Meta `json:"meta"`
}
