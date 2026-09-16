package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestOmitSecretsClearsNestedEncryptedFields(t *testing.T) {
	useDataKey(t)
	data := &AcmeDnsProvider{
		Cloudflare: &AcmeDnsProviderCloudflare{AuthToken: NewEncryptedField("cf-token")},
		Route53:    &AcmeDnsProviderRoute53{SecretAccessKey: NewEncryptedField("r53-key")},
	}

	count, err := OmitSecrets(data)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.True(t, data.Cloudflare.AuthToken.IsEmpty())
	assert.True(t, data.Route53.SecretAccessKey.IsEmpty())
}

// The ciphertext must be gone from what actually gets written, not merely
// cleared on the struct.
func TestOmitSecretsRemovesCiphertextFromTheMarshaledForm(t *testing.T) {
	useDataKey(t)
	setting := &Setting{ID: "s1", Type: base.SettingTypeSecret}
	assert.NoError(t, setting.SetData(&Secret{Key: "DB", Value: NewEncryptedField("hunter2")}))
	assert.True(t, strings.Contains(setting.Data, "hpenc"),
		"precondition: the stored form carries ciphertext")

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	_, err = OmitSecrets(data)
	assert.NoError(t, err)
	assert.NoError(t, setting.SetData(data))

	assert.NotContains(t, setting.Data, "hpenc")
	assert.NotContains(t, setting.Data, "hunter2")
	assert.Contains(t, setting.Data, "DB", "the key stays, so a reader sees a secret is missing")
}

func TestOmitSecretsLeavesNonSecretFieldsAlone(t *testing.T) {
	useDataKey(t)
	data := &RegistryAuth{
		Username: "robot",
		Password: NewEncryptedField("pw"),
		Address:  "registry.example.com",
	}
	count, err := OmitSecrets(data)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "robot", data.Username)
	assert.Equal(t, "registry.example.com", data.Address)
}

func TestOmitSecretsCountsNothingWhenThereAreNone(t *testing.T) {
	useDataKey(t)
	count, err := OmitSecrets(&Script{Data: "echo hi"})
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
