package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetNotificationReq struct {
	settings.GetSettingReq
}

func NewGetNotificationReq() *GetNotificationReq {
	return &GetNotificationReq{}
}

func (req *GetNotificationReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type NotificationResp struct {
	*settings.BaseSettingResp
	ViaEmail        *NotificationViaEmailResp    `json:"viaEmail"`
	ViaSlack        *NotificationViaSlackResp    `json:"viaSlack"`
	ViaDiscord      *NotificationViaDiscordResp  `json:"viaDiscord"`
	ViaTelegram     *NotificationViaTelegramResp `json:"viaTelegram"`
	MinSendInterval timeutil.Duration            `json:"minSendInterval"`
}

type NotificationViaEmailResp struct {
	Enabled          bool                      `json:"enabled"`
	UseDefault       bool                      `json:"useDefault"`
	Sender           *settings.BaseSettingResp `json:"sender"`
	ToProjectMembers bool                      `json:"toProjectMembers"`
	ToProjectOwners  bool                      `json:"toProjectOwners"`
	ToAllAdmins      bool                      `json:"toAllAdmins"`
	ToAddresses      []string                  `json:"toAddresses"`
}

type NotificationViaSlackResp struct {
	Enabled    bool                      `json:"enabled"`
	UseDefault bool                      `json:"useDefault"`
	Webhook    *settings.BaseSettingResp `json:"webhook"`
}

type NotificationViaDiscordResp struct {
	Enabled    bool                      `json:"enabled"`
	UseDefault bool                      `json:"useDefault"`
	Webhook    *settings.BaseSettingResp `json:"webhook"`
}

type NotificationViaTelegramResp struct {
	Enabled    bool                      `json:"enabled"`
	UseDefault bool                      `json:"useDefault"`
	Setting    *settings.BaseSettingResp `json:"setting"`
}

type GetNotificationResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *NotificationResp `json:"data"`
}

func TransformNotification(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *NotificationResp, err error) {
	config := setting.MustAsNotification()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if refObjects == nil {
		refObjects = &entity.RefObjects{}
	}

	viaEmail := resp.ViaEmail
	if viaEmail != nil && viaEmail.Sender != nil && viaEmail.Sender.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaEmail.Sender.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaEmail.Sender.ID, base.SettingTypeEmail)
		}
		viaEmail.Sender = itemResp
	}

	viaSlack := resp.ViaSlack
	if viaSlack != nil && viaSlack.Webhook != nil && viaSlack.Webhook.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaSlack.Webhook.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaSlack.Webhook.ID, base.SettingTypeIMService)
		}
		viaSlack.Webhook = itemResp
	}

	viaDiscord := resp.ViaDiscord
	if viaDiscord != nil && viaDiscord.Webhook != nil && viaDiscord.Webhook.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaDiscord.Webhook.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaDiscord.Webhook.ID, base.SettingTypeIMService)
		}
		viaDiscord.Webhook = itemResp
	}

	viaTelegram := resp.ViaTelegram
	if viaTelegram != nil && viaTelegram.Setting != nil && viaTelegram.Setting.ID != "" {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[viaTelegram.Setting.ID])
		if itemResp == nil {
			itemResp = settings.NewMissingSetting(viaTelegram.Setting.ID, base.SettingTypeIMService)
		}
		viaTelegram.Setting = itemResp
	}

	return resp, nil
}
