package repowebhookdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetRepoWebhookReq struct {
	settings.GetSettingReq
}

func NewGetRepoWebhookReq() *GetRepoWebhookReq {
	return &GetRepoWebhookReq{}
}

func (req *GetRepoWebhookReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRepoWebhookResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *RepoWebhookResp `json:"data"`
}

type RepoWebhookResp struct {
	*settings.BaseSettingResp
	Kind       base.WebhookKind `json:"kind"`
	Secret     string           `json:"secret"`
	WebhookURL string           `json:"webhookURL"`
}

func TransformRepoWebhook(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *RepoWebhookResp, err error) {
	conf := setting.MustAsRepoWebhook()
	if err = copier.Copy(&resp, conf); err != nil {
		return nil, apperrors.New(err)
	}
	resp.Kind = base.WebhookKind(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	// Computed field
	resp.WebhookURL = config.Current.RepoWebhookURL(setting.ID)

	return resp, nil
}
