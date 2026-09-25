package settingmountservice

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func mountOf(source string, parts ...string) *entity.AppSettingMount {
	m := &entity.AppSettingMount{Source: entity.ObjectID{ID: source}}
	for _, part := range parts {
		m.Files = append(m.Files, &entity.AppSettingMountFile{Part: part, Path: "/etc/app/" + part})
	}
	return m
}

func TestGatedByNameWhateverTheType(t *testing.T) {
	for name, want := range map[string]bool{
		"privateKey": true, "password": true, "htpasswd": true,
		"certificate": false, "caCertificate": false, "publicKey": false, "username": false, "other": false,
		"value": false, "content": false,
	} {
		assert.Equal(t, want, GatedPart(name), name)
	}
}

func TestGrantsAreTheSensitivePairsOnce(t *testing.T) {
	assert.Equal(t, []Grant{{Source: "cert_1", Part: "privateKey"}},
		Grants(mountOf("cert_1", "certificate", "privateKey", "privateKey")))
	assert.Empty(t, Grants(mountOf("cert_1", "certificate")))
	assert.Empty(t, Grants(nil))
}

// Renaming an entry, moving a path or changing a mode keeps its pairs: nothing
// new is handed out, so nothing is asked.
func TestWidensOnlyByNewPairs(t *testing.T) {
	before := Grants(mountOf("cert_1", "certificate", "privateKey"))

	assert.Empty(t, Widens(before, Grants(mountOf("cert_1", "privateKey"))), "a part removed")
	assert.Empty(t, Widens(before, before), "unchanged")
	assert.Equal(t, []Grant{{Source: "cert_2", Part: "privateKey"}},
		Widens(before, Grants(mountOf("cert_2", "privateKey"))), "another source")
	assert.Equal(t, []Grant{{Source: "auth_1", Part: "htpasswd"}},
		Widens(nil, Grants(mountOf("auth_1", "username", "htpasswd"))), "a new entry")
}

func TestCheckEntry(t *testing.T) {
	valid := mountOf("cert_1", "certificate", "privateKey")
	assert.NoError(t, CheckEntry("cert", valid, base.SettingTypeSSLCert))

	dup := mountOf("cert_1", "certificate", "certificate")
	samePath := mountOf("cert_1", "certificate", "privateKey")
	samePath.Files[1].Path = samePath.Files[0].Path
	underTLS := mountOf("cert_1", "certificate")
	underTLS.Files[0].Path = "/run/secrets/tls/cert.pem"
	for name, tc := range map[string]struct {
		key   string
		mount *entity.AppSettingMount
		typ   base.SettingType
		want  error
	}{
		"reserved key":   {"tls", valid, base.SettingTypeSSLCert, hperrors.ErrSettingMountKeyInvalid},
		"upper case key": {"Cert", valid, base.SettingTypeSSLCert, hperrors.ErrSettingMountKeyInvalid},
		"no files":       {"cert", mountOf("cert_1"), base.SettingTypeSSLCert, hperrors.ErrSettingMountNoFiles},
		"not a source":   {"cert", valid, base.SettingTypeEmail, hperrors.ErrSettingMountSourceUnsupported},
		"part of other type": {
			"cert", mountOf("cert_1", "htpasswd"), base.SettingTypeSSLCert, hperrors.ErrSettingMountPartInvalid,
		},
		"part twice":     {"cert", dup, base.SettingTypeSSLCert, hperrors.ErrSettingMountPartInvalid},
		"path twice":     {"cert", samePath, base.SettingTypeSSLCert, hperrors.ErrSettingMountPathInvalid},
		"path under tls": {"cert", underTLS, base.SettingTypeSSLCert, hperrors.ErrSettingMountPathInvalid},
	} {
		err := CheckEntry(tc.key, tc.mount, tc.typ)
		assert.True(t, errors.Is(err, tc.want), "%s: %v", name, err)
	}
}

func TestCheckPathsFree(t *testing.T) {
	claimed := map[string]string{"/run/secrets/db_password": "secret DB_PASSWORD"}
	assert.NoError(t, CheckPathsFree(claimed, "/etc/app/key.pem"))
	assert.True(t, errors.Is(CheckPathsFree(claimed, "/run/secrets/db_password"), hperrors.ErrSettingMountPathTaken))
}

func TestFileTargetsOfSecretsAndConfigFiles(t *testing.T) {
	assert.Equal(t, "/run/secrets/db_password", SecretFileTarget(&entity.Secret{
		SwarmRef: &entity.SwarmSecretRef{File: &entity.SwarmRefFileTarget{Name: "db_password"}}}))
	assert.Empty(t, SecretFileTarget(&entity.Secret{}), "read through the environment, no file")
	assert.Equal(t, "/etc/app.conf", ConfigFileTarget(&entity.ConfigFile{
		SwarmRef: &entity.SwarmConfigRef{File: &entity.SwarmRefFileTarget{Name: "/etc/app.conf"}}}))
}
