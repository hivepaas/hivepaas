package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
	var validators []vld.Validator
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
		return nil, apperrors.New(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.New(err)
	}

	if resp.ViaEmail != nil && resp.ViaEmail.Sender != nil {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.ViaEmail.Sender.ID])
		resp.ViaEmail.Sender = itemResp
	}
	if resp.ViaSlack != nil && resp.ViaSlack.Webhook != nil {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.ViaSlack.Webhook.ID])
		resp.ViaSlack.Webhook = itemResp
	}
	if resp.ViaDiscord != nil && resp.ViaDiscord.Webhook != nil {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.ViaDiscord.Webhook.ID])
		resp.ViaDiscord.Webhook = itemResp
	}
	if resp.ViaTelegram != nil && resp.ViaTelegram.Setting != nil {
		itemResp, _ := settings.TransformSettingBase(refObjects.RefSettings[resp.ViaTelegram.Setting.ID])
		resp.ViaTelegram.Setting = itemResp
	}

	return resp, nil
}
