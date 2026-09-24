package entity

import (
	"reflect"
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

// An empty secret is counted wherever it sits; a nested block that is absent
// holds none.
func TestCountEmptySecrets(t *testing.T) {
	assert.Equal(t, 1, CountEmptySecrets(&AppKindSettings{
		Database: &AppKindDatabase{Password: NewEncryptedField("pw")},
	}), "the root password is empty, the password is not; the cache is absent")
	assert.Equal(t, 1, CountEmptySecrets(&Secret{Key: "TOKEN"}))
	assert.Equal(t, 0, CountEmptySecrets(&Secret{Key: "TOKEN", Value: NewEncryptedField("t")}))
	assert.Equal(t, 0, CountEmptySecrets(&AppKindSettings{}))
	assert.Equal(t, 0, CountEmptySecrets(nil))

	assert.Equal(t, 3, countEmptySecretsIn(reflect.ValueOf(&struct {
		List   []Secret
		ByName map[string]*Secret
	}{
		List:   []Secret{{}, {Value: NewEncryptedField("x")}},
		ByName: map[string]*Secret{"a": {}, "b": {}},
	})))
}

func plainOf(t *testing.T, field EncryptedField) string {
	t.Helper()
	plain, err := field.GetPlain()
	assert.NoError(t, err)
	return plain
}

// A setting written without its secrets keeps the ones it replaces, and a
// secret it carries itself wins.
func TestKeepSecrets(t *testing.T) {
	dst := &AppKindSettings{Database: &AppKindDatabase{RootPassword: NewEncryptedField("new-root")}}
	src := &AppKindSettings{Database: &AppKindDatabase{
		Password: NewEncryptedField("old"), RootPassword: NewEncryptedField("old-root"),
	}}

	KeepSecrets(dst, src)

	assert.Equal(t, "old", plainOf(t, dst.Database.Password))
	assert.Equal(t, "new-root", plainOf(t, dst.Database.RootPassword))

	type holder struct {
		List   []Secret
		ByName map[string]*Secret
	}
	listed := &holder{List: []Secret{{}}, ByName: map[string]*Secret{"a": {}, "b": {}}}
	keepSecretsIn(reflect.ValueOf(listed), reflect.ValueOf(&holder{
		List: []Secret{{Value: NewEncryptedField("x")}}, ByName: map[string]*Secret{"a": {Value: NewEncryptedField("y")}},
	}))
	assert.Equal(t, "x", plainOf(t, listed.List[0].Value))
	assert.Equal(t, "y", plainOf(t, listed.ByName["a"].Value))
	assert.True(t, listed.ByName["b"].Value.IsEmpty(), "nothing to keep it from")
}
