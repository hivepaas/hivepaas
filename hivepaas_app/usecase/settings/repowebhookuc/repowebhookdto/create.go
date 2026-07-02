package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	webhookSecretMaxLen = 100
)

type CreateRepoWebhookReq struct {
	settings.CreateSettingReq
	*RepoWebhookBaseReq
}

type RepoWebhookBaseReq struct {
	Name   string           `json:"name"`
	Kind   base.WebhookKind `json:"kind"`
	Secret string           `json:"secret"`
}

func (req *RepoWebhookBaseReq) ToEntity() *entity.RepoWebhook {
	return &entity.RepoWebhook{
		Kind:   req.Kind,
		Secret: req.Secret,
	}
}

func (req *RepoWebhookBaseReq) modifyRequest() error {
	if req.Secret == "" {
		req.Secret = gofn.RandTokenAsHex(base.DefaultWebhookSecretByteLen)
	}
	return nil
}

func (req *RepoWebhookBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Kind, true, base.AllWebhookKinds, field+"kind")...)
	res = append(res, basedto.ValidateStr(&req.Secret, true, 1, webhookSecretMaxLen, field+"secret")...)
	return res
}

func NewCreateRepoWebhookReq() *CreateRepoWebhookReq {
	return &CreateRepoWebhookReq{}
}

func (req *CreateRepoWebhookReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRepoWebhookReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateRepoWebhookResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *RepoWebhookDataResp `json:"data"`
}

type RepoWebhookDataResp struct {
	ID         string `json:"id"`
	Secret     string `json:"secret"`
	WebhookURL string `json:"webhookURL"`
}
