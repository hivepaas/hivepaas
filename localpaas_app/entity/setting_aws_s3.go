package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentAWSS3Version = 1
)

var _ = registerSettingParser(base.SettingTypeAWSS3, &awsS3Parser{})

type awsS3Parser struct {
}

func (s *awsS3Parser) New() SettingData {
	return &AWSS3{}
}

type AWSS3 struct {
	Cred     ObjectID `json:"cred"`
	Region   string   `json:"region,omitempty"`
	Bucket   string   `json:"bucket,omitempty"`
	Endpoint string   `json:"endpoint,omitempty"`
}

func (s *AWSS3) GetType() base.SettingType {
	return base.SettingTypeAWSS3
}

func (s *AWSS3) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s != nil {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Cred.ID)
	}
	return refIDs
}

func (s *AWSS3) MustDecrypt() *AWSS3 {
	return s
}

func (s *AWSS3) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentAWSS3Version == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAWSS3Version
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAWSS3() (*AWSS3, error) {
	return parseSettingAs[*AWSS3](s)
}

func (s *Setting) MustAsAWSS3() *AWSS3 {
	return gofn.Must(s.AsAWSS3())
}
