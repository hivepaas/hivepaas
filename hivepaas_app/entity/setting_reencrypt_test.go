package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// API key secrets are generated with gofn.RandTokenAsHex, so the hashing path
// only accepts hex.
const apiKeySecretHex = "0123456789abcdef0123456789abcdef"

// newStoredSetting builds a setting and returns it the way a row read back from
// the database looks, with no parsed cache.
func newStoredSetting(t *testing.T, data SettingData) *Setting {
	t.Helper()

	useDataKey(t)
	setting := &Setting{ID: "set_1", Type: data.GetType()}
	assert.NoError(t, setting.SetData(data))
	return &Setting{ID: setting.ID, Type: setting.Type, Data: setting.Data}
}

func TestReencryptDataRewritesNestedSecrets(t *testing.T) {
	setting := newStoredSetting(t, &AcmeDnsProvider{
		Cloudflare: &AcmeDnsProviderCloudflare{AuthToken: NewEncryptedField("cf-token")},
		Route53:    &AcmeDnsProviderRoute53{SecretAccessKey: NewEncryptedField("r53-key")},
	})
	before := setting.Data

	changed, err := setting.ReencryptData()
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.NotEqual(t, before, setting.Data, "the ciphertext must actually change")

	reloaded := &Setting{Type: setting.Type, Data: setting.Data}
	provider, err := reloaded.Parse()
	assert.NoError(t, err)

	acme, _ := provider.(*AcmeDnsProvider)
	cfToken, err := acme.Cloudflare.AuthToken.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "cf-token", cfToken)
	r53Key, err := acme.Route53.SecretAccessKey.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "r53-key", r53Key)
}

// A nil provider must not be walked into, and a setting with nothing encrypted
// must not be rewritten at all.
func TestReencryptDataSkipsWhatItShould(t *testing.T) {
	t.Run("no encrypted field means no change", func(t *testing.T) {
		setting := newStoredSetting(t, &ClusterVolume{NodeID: "node_1"})

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("empty data means no change", func(t *testing.T) {
		useDataKey(t)
		setting := &Setting{ID: "set_1", Type: base.SettingTypeClusterVolume}

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("an unset secret is left alone", func(t *testing.T) {
		setting := newStoredSetting(t, &SSHKey{
			PrivateKey: NewEncryptedField("a-private-key"),
			// Passphrase deliberately unset
		})

		changed, err := setting.ReencryptData()
		assert.NoError(t, err)
		assert.True(t, changed)
		assert.NotContains(t, setting.Data, "a-private-key")
	})
}

// The reason the walk is typed: HashField shares the `hpsalt:` prefix but holds a
// hash, so a textual pass over the JSON would try to decrypt it and destroy every
// API key.
func TestReencryptDataLeavesHashedValuesAlone(t *testing.T) {
	useDataKey(t)
	apiKey := &APIKey{KeyID: "key_1", SecretKey: NewHashField(apiKeySecretHex)}
	setting := &Setting{ID: "set_1", Type: apiKey.GetType()}
	assert.NoError(t, setting.SetData(apiKey))

	// Read the stored form back, the way a row loaded from the database looks:
	// in memory HashField still remembers the plaintext it was given.
	reloaded := &Setting{ID: setting.ID, Type: setting.Type, Data: setting.Data}
	hashedBefore := reloaded.MustAsAPIKey().SecretKey.String()
	assert.True(t, strings.HasPrefix(hashedBefore, base.EncryptionSaltPrefix),
		"the stored value should be a hash, got %q", hashedBefore)

	_, err := reloaded.ReencryptData()
	assert.NoError(t, err)

	// The hash must be byte-identical, and must still verify.
	assert.Equal(t, hashedBefore, reloaded.MustAsAPIKey().SecretKey.String())
	assert.NoError(t, reloaded.MustAsAPIKey().SecretKey.VerifyHash(apiKeySecretHex))
}
