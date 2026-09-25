package settingmountservice

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

var testKey = []byte("rotation-key")

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

func sourceSetting(t *testing.T, typ base.SettingType, data entity.SettingData) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "src_1", Type: typ, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func TestEachPartRendersFromItsSource(t *testing.T) {
	useDataKey(t)
	cert := sourceSetting(t, base.SettingTypeSSLCert, &entity.SSLCert{
		Certificate: "CERT", PrivateKey: entity.NewEncryptedField("KEY"), CACertificate: "CA"})
	ssh := sourceSetting(t, base.SettingTypeSSHKey, &entity.SSHKey{
		PublicKey: "PUB", PrivateKey: entity.NewEncryptedField("PRIV")})
	auth := sourceSetting(t, base.SettingTypeBasicAuth, &entity.BasicAuth{
		Username: "admin", Password: entity.NewEncryptedField("s3cret")})

	for _, tc := range []struct {
		setting *entity.Setting
		part    string
		want    string
	}{
		{cert, "certificate", "CERT"}, {cert, "privateKey", "KEY"}, {cert, "caCertificate", "CA"},
		{ssh, "privateKey", "PRIV"}, {ssh, "publicKey", "PUB"},
		{auth, "username", "admin"}, {auth, "password", "s3cret"},
	} {
		values, err := Values(tc.setting)
		assert.NoError(t, err)
		out, err := PartOf(tc.setting.Type, tc.part).RenderFrom(values)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, string(out), "%s %s", tc.setting.Type, tc.part)
	}
}

func TestHtpasswdIsAUsernameAndABcryptOfThePassword(t *testing.T) {
	useDataKey(t)
	auth := sourceSetting(t, base.SettingTypeBasicAuth, &entity.BasicAuth{
		Username: "admin", Password: entity.NewEncryptedField("s3cret")})
	values, err := Values(auth)
	assert.NoError(t, err)

	out, err := PartOf(base.SettingTypeBasicAuth, "htpasswd").RenderFrom(values)
	assert.NoError(t, err)

	user, hash, found := strings.Cut(strings.TrimSuffix(string(out), "\n"), ":")
	assert.True(t, found)
	assert.Equal(t, "admin", user)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("s3cret")))
}

func TestPartsStoredAsSecretsAndPartsGated(t *testing.T) {
	secret, gated := map[string]bool{}, map[string]bool{}
	for _, typ := range SourceTypes() {
		for _, part := range PartsOf(typ) {
			if part.Secret {
				secret[string(typ)+"/"+part.Name] = true
			}
			if part.Gated {
				gated[string(typ)+"/"+part.Name] = true
			}
		}
	}
	assert.Equal(t, map[string]bool{
		"secret/value": true, "ssl-cert/privateKey": true, "ssh-key/privateKey": true,
		"basic-auth/password": true, "basic-auth/htpasswd": true,
	}, secret)
	assert.Equal(t, map[string]bool{
		"ssl-cert/privateKey": true, "ssh-key/privateKey": true,
		"basic-auth/password": true, "basic-auth/htpasswd": true,
	}, gated, "a secret's value is the app's already: mounting it reveals nothing new")
}

// A base64 setting holds bytes; the file is those bytes.
func TestSecretsAndConfigFilesRenderTheirBytes(t *testing.T) {
	useDataKey(t)
	plainSecret := sourceSetting(t, base.SettingTypeSecret, &entity.Secret{
		Key: "DB_PASSWORD", Value: entity.NewEncryptedField("s3cret")})
	binarySecret := sourceSetting(t, base.SettingTypeSecret, &entity.Secret{
		Key: "KEYSTORE", Value: entity.NewEncryptedField("AAEC"), Base64: true})
	plainConfig := sourceSetting(t, base.SettingTypeConfigFile, &entity.ConfigFile{
		Name: "app.conf", Content: "listen 80"})
	binaryConfig := sourceSetting(t, base.SettingTypeConfigFile, &entity.ConfigFile{
		Name: "blob", Content: "AAEC", Base64: true})

	for _, tc := range []struct {
		setting *entity.Setting
		part    string
		want    []byte
	}{
		{plainSecret, "value", []byte("s3cret")},
		{binarySecret, "value", []byte{0, 1, 2}},
		{plainConfig, "content", []byte("listen 80")},
		{binaryConfig, "content", []byte{0, 1, 2}},
	} {
		values, err := Values(tc.setting)
		assert.NoError(t, err)
		out, err := PartOf(tc.setting.Type, tc.part).RenderFrom(values)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, out, "%s %s", tc.setting.Type, tc.part)
	}
}

// bcrypt salts every hash afresh, so the key reads the inputs, never the output:
// comparing output would restart the app on every refresh.
func TestTheRotationKeyFollowsInputsAndVersionOnly(t *testing.T) {
	part := PartOf(base.SettingTypeBasicAuth, "htpasswd")
	values := map[string]string{"username": "admin", "password": "s3cret"}
	key := RotationKey(testKey, base.SettingTypeBasicAuth, part, values)

	assert.Equal(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, part, values), "stable")
	assert.NotEqual(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, part,
		map[string]string{"username": "admin", "password": "other"}), "an input changed")
	bumped := *part
	bumped.Version++
	assert.NotEqual(t, key, RotationKey(testKey, base.SettingTypeBasicAuth, &bumped, values), "the version changed")
	assert.NotEqual(t, key, RotationKey([]byte("another-key"), base.SettingTypeBasicAuth, part, values),
		"keyed: a name says nothing of a password to whoever lists secrets")
	// Inputs are length-prefixed: moving a character from one to the next is a change.
	assert.NotEqual(t,
		RotationKey(testKey, base.SettingTypeBasicAuth, part, map[string]string{"username": "ab", "password": "c"}),
		RotationKey(testKey, base.SettingTypeBasicAuth, part, map[string]string{"username": "a", "password": "bc"}))
}

func TestASourceIsUsableWhenEveryRequiredPartHasItsInputs(t *testing.T) {
	assert.True(t, Usable(base.SettingTypeSSLCert, map[string]string{"certificate": "C", "privateKey": "K"}))
	assert.False(t, Usable(base.SettingTypeSSLCert, map[string]string{"certificate": "", "privateKey": "K"}),
		"a certificate not obtained yet")
	assert.False(t, Usable(base.SettingTypeEmail, map[string]string{}), "not a source type")
}
