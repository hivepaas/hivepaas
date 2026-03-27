package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
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
	Provider ObjectID        `json:"provider"`
	S3       *CloudStorageS3 `json:"s3,omitempty"`

	// NOTE: this is a temporary field
	RefProvider *CloudProvider `json:"-"`
}

type CloudStorageS3 struct {
	Region   string `json:"region,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

func (s *CloudStorage) GetType() base.SettingType {
	return base.SettingTypeCloudStorage
}

func (s *CloudStorage) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Provider.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Provider.ID)
	}
	return refIDs
}

func (s *CloudStorage) MustDecrypt() *CloudStorage {
	return s
}

func (s *CloudStorage) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentCloudStorageVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCloudStorageVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsCloudStorage() (*CloudStorage, error) {
	return parseSettingAs[*CloudStorage](s)
}

func (s *Setting) MustAsCloudStorage() *CloudStorage {
	return gofn.Must(s.AsCloudStorage())
}
