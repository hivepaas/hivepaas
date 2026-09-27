package entity

import (
	"github.com/tiendc/gofn"

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
	// KeyAuth is the key auth setting whose key id and secret key the bucket is
	// reached with.
	KeyAuth  ObjectID `json:"keyAuth"`
	Region   string   `json:"region,omitempty"`
	Bucket   string   `json:"bucket,omitempty"`
	Endpoint string   `json:"endpoint,omitempty"`
}

func (s *CloudStorage) GetType() base.SettingType {
	return base.SettingTypeCloudStorage
}

func (s *CloudStorage) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.S3 != nil && s.S3.KeyAuth.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.S3.KeyAuth.ID)
	}
	return refIDs
}

func (s *CloudStorage) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

// Decrypt has nothing to do: the secret is the linked key auth's.
func (s *CloudStorage) Decrypt() error {
	return nil
}

func (s *Setting) AsCloudStorage() (*CloudStorage, error) {
	return parseSettingAs[*CloudStorage](s)
}

func (s *Setting) MustAsCloudStorage() *CloudStorage {
	return gofn.Must(s.AsCloudStorage())
}
