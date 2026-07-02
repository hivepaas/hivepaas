package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentIMServiceVersion = 1
)

var _ = registerSettingParser(base.SettingTypeIMService, &imServiceParser{})

type imServiceParser struct {
}

func (s *imServiceParser) New() SettingData {
	return &IMService{}
}

type IMService struct {
	Slack    *IMSlack    `json:"slack,omitempty"`
	Discord  *IMDiscord  `json:"discord,omitempty"`
	Telegram *IMTelegram `json:"telegram,omitempty"`
}

type IMSlack struct {
	Webhook EncryptedField `json:"webhook"`
}

type IMDiscord struct {
	Webhook EncryptedField `json:"webhook"`
}

type IMTelegram struct {
	BotToken EncryptedField `json:"botToken"`
	ChatID   string         `json:"chatId"`
}

func (s *IMService) GetType() base.SettingType {
	return base.SettingTypeIMService
}

func (s *IMService) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *IMService) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *IMService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentIMServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentIMServiceVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentIMServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *IMService) Decrypt() error {
	if s.Slack != nil {
		_, err := s.Slack.Webhook.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
	}
	if s.Discord != nil {
		_, err := s.Discord.Webhook.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
	}
	if s.Telegram != nil {
		_, err := s.Telegram.BotToken.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
	}
	return nil
}

func (s *Setting) AsIMService() (*IMService, error) {
	return parseSettingAs[*IMService](s)
}

func (s *Setting) MustAsIMService() *IMService {
	return gofn.Must(s.AsIMService())
}
