package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
		Secret: entity.NewEncryptedField(req.Secret),
	}
}

// IsSecretMasked reports whether the request carries the placeholder the GET
// response substitutes for a stored secret, rather than a real value. A form that
// posts the response back unchanged must not overwrite the secret with the mask.
func (req *RepoWebhookBaseReq) IsSecretMasked() bool {
	return req.Secret == maskedSecret
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
	res = append(res, basedto.ValidateStrIn(&req.Kind, false, base.AllWebhookKinds, field+"kind")...)
	res = append(res, basedto.ValidateStr(&req.Secret, true, 1, webhookSecretMaxLen, field+"secret")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Secret, field+"secret")...)
	return res
}

func NewCreateRepoWebhookReq() *CreateRepoWebhookReq {
	return &CreateRepoWebhookReq{}
}

func (req *CreateRepoWebhookReq) ModifyRequest() error {
	return req.modifyRequest()
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRepoWebhookReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
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
