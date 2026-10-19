package managedserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
)

type fakeSettings struct {
	repository.SettingRepo
	byID map[string]*entity.Setting
}

func (f *fakeSettings) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, typ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if setting := f.byID[id]; setting != nil && setting.Type == typ {
		return setting, nil
	}
	return nil, hperrors.NewNotFound("Setting")
}

func (f *fakeSettings) add(id string, typ base.SettingType, kind string, data entity.SettingData) {
	setting := &entity.Setting{ID: id, Type: typ, Kind: kind, Status: base.SettingStatusActive}
	setting.MustSetData(data)
	f.byID[id] = setting
}

func newResolveTest(t *testing.T) *service {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	settings := &fakeSettings{byID: map[string]*entity.Setting{}}
	settings.add("ka-1", base.SettingTypeKeyAuth, "", &entity.KeyAuth{KeyID: "AKIA-STORAGE",
		SecretKey: entity.NewEncryptedField("storage-secret")})
	settings.add("ka-narrow", base.SettingTypeKeyAuth, "", &entity.KeyAuth{KeyID: "AKIA-NARROW",
		SecretKey: entity.NewEncryptedField("narrow-secret")})
	settings.add("cs-1", base.SettingTypeCloudStorage, string(base.CloudStorageKindS3), &entity.CloudStorage{
		S3: &entity.CloudStorageS3{KeyAuth: entity.ObjectID{ID: "ka-1"}, Region: "auto", Bucket: "backups",
			Endpoint: "https://s3.example.com"}})
	settings.add("repo-s3", base.SettingTypeBackupRepo, "", &entity.BackupRepo{
		CloudStorage: entity.ObjectID{ID: "cs-1"}, StoragePrefix: "kopia"})
	settings.add("repo-vol", base.SettingTypeBackupRepo, "", &entity.BackupRepo{Volume: entity.ObjectID{ID: "v"}})
	return &service{settingRepo: settings}
}

func TestResolveS3ThroughARepository(t *testing.T) {
	svc := newResolveTest(t)

	s3, err := svc.ResolveS3(context.Background(), nil, nil, entity.ManagedStorage{
		BackupRepo: entity.ObjectID{ID: "repo-s3"}, Prefix: "kopia/walg/db"})

	assert.NoError(t, err)
	assert.Equal(t, "backups", s3.Bucket)
	assert.Equal(t, "kopia/walg/db", s3.Prefix)
	assert.Equal(t, "AKIA-STORAGE", s3.AccessKey)
	assert.Equal(t, "storage-secret", s3.SecretKey)
	assert.True(t, s3.ForcePathStyle)
}

func TestResolveS3WithANarrowerKey(t *testing.T) {
	svc := newResolveTest(t)

	s3, err := svc.ResolveS3(context.Background(), nil, nil, entity.ManagedStorage{
		CloudStorage: entity.ObjectID{ID: "cs-1"}, KeyAuth: entity.ObjectID{ID: "ka-narrow"}, Prefix: "walg/db"})

	assert.NoError(t, err)
	assert.Equal(t, "AKIA-NARROW", s3.AccessKey)
	assert.Equal(t, "narrow-secret", s3.SecretKey)
}

func TestResolveS3OfAVolumeRepositoryIsRefused(t *testing.T) {
	svc := newResolveTest(t)

	_, err := svc.ResolveS3(context.Background(), nil, nil, entity.ManagedStorage{
		BackupRepo: entity.ObjectID{ID: "repo-vol"}})

	assert.ErrorIs(t, err, hperrors.ErrStorageTypeUnsupported)
}
