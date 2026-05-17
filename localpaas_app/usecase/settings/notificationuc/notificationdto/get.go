package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
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
	ViaEmail        *NotificationViaEmailResp   `json:"viaEmail"`
	ViaSlack        *NotificationViaSlackResp   `json:"viaSlack"`
	ViaDiscord      *NotificationViaDiscordResp `json:"viaDiscord"`
	MinSendInterval timeutil.Duration           `json:"minSendInterval"`
}

type NotificationViaEmailResp struct {
	Enabled          bool                      `json:"enabled"`
	Sender           *settings.BaseSettingResp `json:"sender"`
	ToProjectMembers bool                      `json:"toProjectMembers"`
	ToProjectOwners  bool                      `json:"toProjectOwners"`
	ToAllAdmins      bool                      `json:"toAllAdmins"`
	ToAddresses      []string                  `json:"toAddresses"`
}

type NotificationViaSlackResp struct {
	Enabled bool                      `json:"enabled"`
	Webhook *settings.BaseSettingResp `json:"webhook"`
}

type NotificationViaDiscordResp struct {
	Enabled bool                      `json:"enabled"`
	Webhook *settings.BaseSettingResp `json:"webhook"`
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

	return resp, nil
}
