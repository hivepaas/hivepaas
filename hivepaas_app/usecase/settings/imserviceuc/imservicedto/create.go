package imservicedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	webhookURLMaxLen = 512
	tokenMaxLen      = 200
)

type CreateIMServiceReq struct {
	settings.CreateSettingReq
	*IMServiceBaseReq
}

type IMServiceBaseReq struct {
	Name     string             `json:"name"`
	Kind     base.IMServiceKind `json:"kind"`
	Slack    *IMSlackReq        `json:"slack"`
	Discord  *IMDiscordReq      `json:"discord"`
	Telegram *IMTelegramReq     `json:"telegram"`
	Lark     *IMLarkReq         `json:"lark"`
}

func (req *IMServiceBaseReq) ToEntity() *entity.IMService {
	imService := &entity.IMService{}
	switch req.Kind {
	case base.IMServiceKindSlack:
		imService.Slack = req.Slack.ToEntity()
	case base.IMServiceKindDiscord:
		imService.Discord = req.Discord.ToEntity()
	case base.IMServiceKindTelegram:
		imService.Telegram = req.Telegram.ToEntity()
	case base.IMServiceKindLark:
		imService.Lark = req.Lark.ToEntity()
	}
	return imService
}

type IMSlackReq struct {
	Webhook string `json:"webhook"`
}

func (req *IMSlackReq) ToEntity() *entity.IMSlack {
	return &entity.IMSlack{
		Webhook: entity.NewEncryptedField(req.Webhook),
	}
}

func (req *IMSlackReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Webhook, true, 1, webhookURLMaxLen, field+"webhook")...)
	return res
}

type IMDiscordReq struct {
	Webhook string `json:"webhook"`
}

func (req *IMDiscordReq) ToEntity() *entity.IMDiscord {
	return &entity.IMDiscord{
		Webhook: entity.NewEncryptedField(req.Webhook),
	}
}

func (req *IMDiscordReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Webhook, true, 1, webhookURLMaxLen, field+"webhook")...)
	return res
}

type IMTelegramReq struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}

func (req *IMTelegramReq) ToEntity() *entity.IMTelegram {
	return &entity.IMTelegram{
		BotToken: entity.NewEncryptedField(req.BotToken),
		ChatID:   req.ChatID,
	}
}

func (req *IMTelegramReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.BotToken, true, 1, tokenMaxLen, field+"botToken")...)
	res = append(res, basedto.ValidateStr(&req.ChatID, true, 1, tokenMaxLen, field+"chatId")...)
	return res
}

type IMLarkReq struct {
	Webhook string `json:"webhook"`
	Secret  string `json:"secret,omitempty"`
}

func (req *IMLarkReq) ToEntity() *entity.IMLark {
	return &entity.IMLark{
		Webhook: entity.NewEncryptedField(req.Webhook),
		Secret:  entity.NewEncryptedField(req.Secret),
	}
}

func (req *IMLarkReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return nil
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Webhook, true, 1, webhookURLMaxLen, field+"webhook")...)
	res = append(res, basedto.ValidateStr(&req.Secret, false, 0, tokenMaxLen, field+"secret")...)
	return res
}

func (req *IMServiceBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	switch req.Kind {
	case base.IMServiceKindSlack:
		res = append(res, basedto.ValidateCond(req.Slack != nil, field+"slack")...)
		res = append(res, req.Slack.validate(field+"slack")...)
	case base.IMServiceKindDiscord:
		res = append(res, basedto.ValidateCond(req.Discord != nil, field+"discord")...)
		res = append(res, req.Discord.validate(field+"discord")...)
	case base.IMServiceKindTelegram:
		res = append(res, basedto.ValidateCond(req.Telegram != nil, field+"telegram")...)
		res = append(res, req.Telegram.validate(field+"telegram")...)
	case base.IMServiceKindLark:
		res = append(res, basedto.ValidateCond(req.Lark != nil, field+"lark")...)
		res = append(res, req.Lark.validate(field+"lark")...)
	}
	return res
}

func NewCreateIMServiceReq() *CreateIMServiceReq {
	return &CreateIMServiceReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateIMServiceReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateIMServiceResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
