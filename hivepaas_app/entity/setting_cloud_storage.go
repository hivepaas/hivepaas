package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentCloudStorageVersion = 1
)

var _ = registerSettingParser(base.SettingTypeCloudStorage, &cloudStorageParser{})

type cloudStorageParser struct {
}

func (s *cloudStorageParser) New() SettingData {
	return &CloudStorage{}
}

type CloudStorage struct {
	S3 *CloudStorageS3 `json:"s3,omitempty"`
}

type CloudStorageS3 struct {
	*CloudProviderAWS
	Region   string `json:"region,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

func (s *CloudStorage) GetType() base.SettingType {
	return base.SettingTypeCloudStorage
}

func (s *CloudStorage) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *CloudStorage) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *CloudStorage) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCloudStorageVersion {
		return false, nil
	}
	if setting.Version > CurrentCloudStorageVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCloudStorageVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *CloudStorage) Decrypt() error {
	if s.S3 != nil {
		_, err := s.S3.SecretKey.GetPlain()
		if err != nil {
			return apperrors.Wrap(err)
		}
	}
	return nil
}
func (s *Setting) AsCloudStorage() (*CloudStorage, error) {
	return parseSettingAs[*CloudStorage](s)
}

func (s *Setting) MustAsCloudStorage() *CloudStorage {
	return gofn.Must(s.AsCloudStorage())
}
