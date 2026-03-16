package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
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
	Slack   *Slack   `json:"slack,omitempty"`
	Discord *Discord `json:"discord,omitempty"`
}

type Slack struct {
	Webhook EncryptedField `json:"webhook"`
}

type Discord struct {
	Webhook EncryptedField `json:"webhook"`
}

func (s *IMService) GetType() base.SettingType {
	return base.SettingTypeIMService
}

func (s *IMService) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *IMService) MustDecrypt() *IMService {
	if s.Slack != nil {
		s.Slack.Webhook.MustGetPlain()
	}
	if s.Discord != nil {
		s.Discord.Webhook.MustGetPlain()
	}
	return s
}

func (s *IMService) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentIMServiceVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentIMServiceVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsIMService() (*IMService, error) {
	return parseSettingAs[*IMService](s)
}

func (s *Setting) MustAsIMService() *IMService {
	return gofn.Must(s.AsIMService())
}
