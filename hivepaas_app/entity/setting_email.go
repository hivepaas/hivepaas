package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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
	Method       base.HTTPMethod        `json:"method"`
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

func (s *Email) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Email) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentEmailVersion {
		return false, nil
	}
	if setting.Version > CurrentEmailVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentEmailVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Email) Decrypt() error {
	if s.SMTP != nil {
		_, err := s.SMTP.Password.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
	}
	if s.HTTP != nil {
		_, err := s.HTTP.Password.GetPlain()
		if err != nil {
			return apperrors.New(err)
		}
	}
	return nil
}
func (s *Setting) AsEmail() (*Email, error) {
	return parseSettingAs[*Email](s)
}

func (s *Setting) MustAsEmail() *Email {
	return gofn.Must(s.AsEmail())
}
