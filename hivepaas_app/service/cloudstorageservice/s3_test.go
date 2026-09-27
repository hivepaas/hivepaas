package cloudstorageservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

func storageSetting(t *testing.T, keyAuthID string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "cs1", Type: base.SettingTypeCloudStorage, Kind: string(base.CloudStorageKindS3)}
	assert.NoError(t, setting.SetData(&entity.CloudStorage{S3: &entity.CloudStorageS3{
		KeyAuth: entity.ObjectID{ID: keyAuthID}, Region: "eu-west-1", Bucket: "backups",
		Endpoint: "https://s3.example.test",
	}}))
	return setting
}

func keyAuthSetting(t *testing.T, status base.SettingStatus) *entity.Setting {
	t.Helper()
	// Storing a secret takes the data key, as it does at runtime.
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	setting := &entity.Setting{ID: "ka1", Type: base.SettingTypeKeyAuth, Name: "aws", Status: status}
	assert.NoError(t, setting.SetData(&entity.KeyAuth{KeyID: "AKIAEXAMPLE",
		SecretKey: entity.NewEncryptedField("s3cr3t")}))
	return setting
}

// A storage is its bucket and the key of the key auth it links.
func TestS3ConfigTakesTheKeyOfTheLinkedKeyAuth(t *testing.T) {
	refs := entity.NewRefObjects()
	refs.RefSettings["ka1"] = keyAuthSetting(t, base.SettingStatusActive)

	cfg, err := S3Config(context.Background(), nil, storageSetting(t, "ka1"), refs)
	if assert.NoError(t, err) {
		assert.Equal(t, "AKIAEXAMPLE", cfg.AccessKeyID)
		assert.Equal(t, "s3cr3t", cfg.SecretAccessKey)
		assert.Equal(t, "backups", cfg.Bucket)
		assert.Equal(t, "eu-west-1", cfg.Region)
		assert.Equal(t, "https://s3.example.test", cfg.Endpoint)
	}
}

// A disabled key auth, or none, is a storage that cannot be used: said here,
// rather than by S3 later.
func TestS3ConfigRefusesAKeyAuthNotThere(t *testing.T) {
	refs := entity.NewRefObjects()
	refs.RefSettings["ka1"] = keyAuthSetting(t, base.SettingStatusDisabled)

	_, err := S3Config(context.Background(), nil, storageSetting(t, "ka1"), refs)
	assert.ErrorIs(t, err, hperrors.ErrSettingNotFound)

	_, err = S3Config(context.Background(), nil, storageSetting(t, ""), refs)
	assert.ErrorIs(t, err, hperrors.ErrSettingNotFound)
}

func TestACloudStorageReferencesItsKeyAuth(t *testing.T) {
	storage := &entity.CloudStorage{S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "ka1"}}}
	assert.Equal(t, []string{"ka1"}, storage.GetRefObjectIDs().RefSettingIDs)
}
