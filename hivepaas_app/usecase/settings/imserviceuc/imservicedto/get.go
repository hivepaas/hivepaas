package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maskedSecret = "****************"
)

type GetIMServiceReq struct {
	settings.GetSettingReq
}

func NewGetIMServiceReq() *GetIMServiceReq {
	return &GetIMServiceReq{}
}

func (req *GetIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetIMServiceResp struct {
	Meta *basedto.Meta  `json:"meta"`
	Data *IMServiceResp `json:"data"`
}

type IMServiceResp struct {
	*settings.BaseSettingResp
	Kind         base.IMServiceKind `json:"kind"`
	Slack        *IMSlackResp       `json:"slack,omitempty"`
	Discord      *IMDiscordResp     `json:"discord,omitempty"`
	Telegram     *IMTelegramResp    `json:"telegram,omitempty"`
	Lark         *IMLarkResp        `json:"lark,omitempty"`
	SecretMasked bool               `json:"secretMasked,omitempty"`
}

type IMSlackResp struct {
	Webhook string `json:"webhook"`
}

func (resp *IMSlackResp) CopyWebhook(field entity.EncryptedField) error {
	resp.Webhook = field.String()
	return nil
}

type IMDiscordResp struct {
	Webhook string `json:"webhook"`
}

func (resp *IMDiscordResp) CopyWebhook(field entity.EncryptedField) error {
	resp.Webhook = field.String()
	return nil
}

type IMTelegramResp struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}

func (resp *IMTelegramResp) CopyBotToken(field entity.EncryptedField) error {
	resp.BotToken = field.String()
	return nil
}

type IMLarkResp struct {
	Webhook string `json:"webhook"`
	Secret  string `json:"secret,omitempty"`
}

func (resp *IMLarkResp) CopyWebhook(field entity.EncryptedField) error {
	resp.Webhook = field.String()
	return nil
}

func (resp *IMLarkResp) CopySecret(field entity.EncryptedField) error {
	resp.Secret = field.String()
	return nil
}

func TransformIMService(
	setting *entity.Setting,
	_ *entity.RefObjects,
) (resp *IMServiceResp, err error) {
	config := setting.MustAsIMService()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Kind = base.IMServiceKind(setting.Kind)

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	switch {
	case config.Slack != nil:
		resp.SecretMasked = config.Slack.Webhook.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Slack.Webhook = maskedSecret
		}
	case config.Discord != nil:
		resp.SecretMasked = config.Discord.Webhook.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Discord.Webhook = maskedSecret
		}
	case config.Telegram != nil:
		resp.SecretMasked = config.Telegram.BotToken.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Telegram.BotToken = maskedSecret
		}
	case config.Lark != nil:
		resp.SecretMasked = config.Lark.Webhook.IsEncrypted() || config.Lark.Secret.IsEncrypted() || resp.Inherited
		if resp.SecretMasked {
			resp.Lark.Webhook = maskedSecret
			if config.Lark.Secret.String() != "" {
				resp.Lark.Secret = maskedSecret
			}
		}
	}

	return resp, nil
}
