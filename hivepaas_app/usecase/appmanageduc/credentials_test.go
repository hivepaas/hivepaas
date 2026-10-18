package appmanageduc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

func secretSetting(t *testing.T, id, key, value string) *entity.Setting {
	t.Helper()
	s := &entity.Setting{ID: id, Type: base.SettingTypeSecret, Name: key, UpdateVer: 1}
	s.MustSetData(&entity.Secret{Key: key, Value: entity.NewEncryptedField(value)})
	return s
}

// Refreshing rewrites the secrets whose value the storage's key changed, and
// only those: one that did not change is not mounted again.
func TestRefreshRewritesTheSecretsThatChanged(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	access := secretSetting(t, "s1", "WALG_AWS_ACCESS_KEY_ID", "AKIA-OLD")
	secret := secretSetting(t, "s2", "WALG_AWS_SECRET_ACCESS_KEY", "same-secret")
	sodium := secretSetting(t, "s3", "WALG_LIBSODIUM_KEY", "never-touched")

	changed, err := refreshedSecrets([]*entity.Setting{access, secret, sodium}, map[string]string{
		"WALG_AWS_ACCESS_KEY_ID": "AKIA-NEW", "WALG_AWS_SECRET_ACCESS_KEY": "same-secret"})

	assert.NoError(t, err)
	assert.Equal(t, []*entity.Setting{access}, changed)
	assert.Equal(t, 2, access.UpdateVer)
	stored := &entity.Setting{Type: base.SettingTypeSecret, Data: access.Data}
	data, err := stored.AsSecret()
	assert.NoError(t, err)
	plain, _ := data.Value.GetPlain()
	assert.Equal(t, "AKIA-NEW", plain)
	assert.Equal(t, 1, sodium.UpdateVer)
}
