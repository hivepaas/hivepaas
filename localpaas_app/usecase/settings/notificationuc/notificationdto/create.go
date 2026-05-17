package notificationdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
)

type CreateNotificationReq struct {
	settings.CreateSettingReq
	*NotificationBaseReq
}

type NotificationBaseReq struct {
	Name            string                     `json:"name"`
	ViaEmail        *NotificationViaEmailReq   `json:"viaEmail"`
	ViaSlack        *NotificationViaSlackReq   `json:"viaSlack"`
	ViaDiscord      *NotificationViaDiscordReq `json:"viaDiscord"`
	MinSendInterval timeutil.Duration          `json:"minSendInterval"`
}

func (req *NotificationBaseReq) ToEntity() *entity.Notification {
	return &entity.Notification{
		ViaEmail:        req.ViaEmail.ToEntity(),
		ViaSlack:        req.ViaSlack.ToEntity(),
		ViaDiscord:      req.ViaDiscord.ToEntity(),
		MinSendInterval: req.MinSendInterval,
	}
}

type NotificationViaEmailReq struct {
	Enabled          bool                `json:"enabled"`
	Sender           basedto.ObjectIDReq `json:"sender"`
	ToProjectMembers bool                `json:"toProjectMembers"`
	ToProjectOwners  bool                `json:"toProjectOwners"`
	ToAllAdmins      bool                `json:"toAllAdmins"`
	ToAddresses      []string            `json:"toAddresses"`
}

func (req *NotificationViaEmailReq) ToEntity() *entity.NotificationViaEmail {
	if req == nil {
		return nil
	}
	return &entity.NotificationViaEmail{
		Enabled:          req.Enabled,
		Sender:           entity.ObjectID{ID: req.Sender.ID},
		ToProjectMembers: req.ToProjectMembers,
		ToProjectOwners:  req.ToProjectOwners,
		ToAllAdmins:      req.ToAllAdmins,
		ToAddresses:      req.ToAddresses,
	}
}

type NotificationViaSlackReq struct {
	Enabled bool                `json:"enabled"`
	Webhook basedto.ObjectIDReq `json:"webhook"`
}

func (req *NotificationViaSlackReq) ToEntity() *entity.NotificationViaSlack {
	if req == nil {
		return nil
	}
	return &entity.NotificationViaSlack{
		Enabled: req.Enabled,
		Webhook: entity.ObjectID{ID: req.Webhook.ID},
	}
}

type NotificationViaDiscordReq struct {
	Enabled bool                `json:"enabled"`
	Webhook basedto.ObjectIDReq `json:"webhook"`
}

func (req *NotificationViaDiscordReq) ToEntity() *entity.NotificationViaDiscord {
	if req == nil {
		return nil
	}
	return &entity.NotificationViaDiscord{
		Enabled: req.Enabled,
		Webhook: entity.ObjectID{ID: req.Webhook.ID},
	}
}

func (req *NotificationBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	return res
}

func NewCreateNotificationReq() *CreateNotificationReq {
	return &CreateNotificationReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateNotificationReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateNotificationResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
