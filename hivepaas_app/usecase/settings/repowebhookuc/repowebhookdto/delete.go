package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type DeleteRepoWebhookReq struct {
	settings.DeleteSettingReq
}

func NewDeleteRepoWebhookReq() *DeleteRepoWebhookReq {
	return &DeleteRepoWebhookReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteRepoWebhookReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.DeleteSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type DeleteRepoWebhookResp struct {
	Meta *basedto.Meta `json:"meta"`
}
