package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentEmailVersion = 1
)

var _ = registerSettingParser(base.SettingTypeEmail, &emailParser{})

type emailParser struct {
}

func (s *emailParser) New() SettingData {
	return &Email{}
}

type Email struct {
	SMTP *EmailSMTP `json:"smtp,omitempty"`
	HTTP *EmailHTTP `json:"http,omitempty"`
}

type EmailSMTP struct {
	Host        string         `json:"host"`
	Port        int            `json:"port"`
	Username    string         `json:"username"`
	DisplayName string         `json:"displayName"`
	Password    EncryptedField `json:"password"`
	SSL         bool           `json:"ssl"`
}

type EmailHTTP struct {
	Endpoint     string                 `json:"endpoint"`
	Method       string                 `json:"method"`
	ContentType  string                 `json:"contentType"`
	Headers      map[string]string      `json:"headers"`
	FieldMapping *EmailHTTPFieldMapping `json:"fieldMapping"`
	Username     string                 `json:"username"`
	DisplayName  string                 `json:"displayName"`
	Password     EncryptedField         `json:"password"`
}

type EmailHTTPFieldMapping struct {
	FromAddress string `json:"fromAddress"`
	FromName    string `json:"fromName"`
	ToAddress   string `json:"toAddress"`
	ToAddresses string `json:"toAddresses"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	Password    string `json:"password"`
}

func (s *Email) GetType() base.SettingType {
	return base.SettingTypeEmail
}

func (s *Email) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *Email) MustDecrypt() *Email {
	if s.SMTP != nil {
		s.SMTP.Password.MustGetPlain()
	}
	if s.HTTP != nil {
		s.HTTP.Password.MustGetPlain()
	}
	return s
}

func (s *Email) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentEmailVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentEmailVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsEmail() (*Email, error) {
	return parseSettingAs[*Email](s)
}

func (s *Setting) MustAsEmail() *Email {
	return gofn.Must(s.AsEmail())
}
