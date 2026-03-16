package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentAWSVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAWS, &awsParser{})

type awsParser struct {
}

func (s *awsParser) New() SettingData {
	return &AWS{}
}

type AWS struct {
	AccessKeyID string         `json:"accessKeyId"`
	SecretKey   EncryptedField `json:"secretKey"`
	Region      string         `json:"region,omitempty"`
}

func (s *AWS) GetType() base.SettingType {
	return base.SettingTypeAWS
}

func (s *AWS) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AWS) MustDecrypt() *AWS {
	s.SecretKey.MustGetPlain()
	return s
}

func (s *AWS) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentAWSVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAWSVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAWS() (*AWS, error) {
	return parseSettingAs[*AWS](s)
}

func (s *Setting) MustAsAWS() *AWS {
	return gofn.Must(s.AsAWS())
}
