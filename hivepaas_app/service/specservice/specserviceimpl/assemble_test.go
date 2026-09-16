package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func useDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
}

func mustYAML(t *testing.T, v any) string {
	t.Helper()
	out, err := yaml.Marshal(v)
	assert.NoError(t, err)
	return string(out)
}

func settingWith(t *testing.T, s *entity.Setting, data entity.SettingData) *entity.Setting {
	t.Helper()
	assert.NoError(t, s.SetData(data))
	return s
}

func TestAssembleSettingsPutsSingletonsUnderTheirBlockName(t *testing.T) {
	useDataKey(t)
	routing := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting},
		&entity.AppRoutingSettings{Port: 8080})

	out, err := assembleSettings([]*entity.Setting{routing}, newRefIndex(), specmodel.SecretsModePlaintext)
	assert.NoError(t, err)

	block, ok := out["routing"]
	assert.True(t, ok, "a singleton takes its block name, with no key")
	assert.NotNil(t, block)
	assert.NotContains(t, out, "routings")
}

func TestAssembleSettingsPutsCollectionsInAKeyedMap(t *testing.T) {
	useDataKey(t)
	cert := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeSSLCert, Name: "localhost", Kind: "self-signed"},
		&entity.SSLCert{Domain: "localhost"})

	out, err := assembleSettings([]*entity.Setting{cert}, newRefIndex(), specmodel.SecretsModePlaintext)
	assert.NoError(t, err)

	certs, ok := out["sslCerts"].(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, certs, "localhost", "keyed by the derived key, not by id")
}

// A reference whose target is in the export becomes a path carrying its scope.
func TestAssembleSettingsRewritesInScopeReferencesToPaths(t *testing.T) {
	useDataKey(t)
	index := newRefIndex()
	index.addPath("cert_1", "global/sslCerts/localhost@self-signed")

	routing := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting},
		&entity.AppRoutingSettings{
			Port:    443,
			Domains: []*entity.AppDomain{{Domain: "x.com", SSLCert: entity.ObjectID{ID: "cert_1"}}},
		})

	out, err := assembleSettings([]*entity.Setting{routing}, index, specmodel.SecretsModePlaintext)
	assert.NoError(t, err)

	rendered := mustYAML(t, out)
	assert.Contains(t, rendered, "global/sslCerts/localhost@self-signed")
	assert.NotContains(t, rendered, "cert_1")
}

// A reference outside the exported scope becomes an external block with enough
// to find the target on the importing installation.
func TestAssembleSettingsWritesExternalRefsForOutOfScopeTargets(t *testing.T) {
	useDataKey(t)
	index := newRefIndex()
	index.addExternal(&entity.Setting{
		ID: "cert_9", Type: base.SettingTypeSSLCert, Name: "wildcard", Kind: "letsencrypt",
	})

	routing := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting},
		&entity.AppRoutingSettings{
			Domains: []*entity.AppDomain{{Domain: "x.com", SSLCert: entity.ObjectID{ID: "cert_9"}}},
		})

	out, err := assembleSettings([]*entity.Setting{routing}, index, specmodel.SecretsModePlaintext)
	assert.NoError(t, err)

	rendered := mustYAML(t, out)
	assert.Contains(t, rendered, "external:")
	assert.Contains(t, rendered, "wildcard")
	assert.Contains(t, rendered, "letsencrypt")
	assert.Contains(t, rendered, "cert_9", "the source id travels so a same-system import matches exactly")
}

// The policy's Strip runs during assembly, or a transient flag reaches the file.
func TestAssembleSettingsAppliesTheTypePolicyStrip(t *testing.T) {
	useDataKey(t)
	routing := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting},
		&entity.AppRoutingSettings{Port: 8080, Reset: true})

	out, err := assembleSettings([]*entity.Setting{routing}, newRefIndex(), specmodel.SecretsModePlaintext)
	assert.NoError(t, err)
	assert.NotContains(t, mustYAML(t, out), "reset",
		"Reset is a command and importing it would perform a reset")
}

// A type in neither block-name map is a programming error, not a silent skip.
func TestAssembleSettingsRefusesAnUnnamedType(t *testing.T) {
	_, err := assembleSettings([]*entity.Setting{
		{ID: "s1", Type: base.SettingType("brand-new-type")},
	}, newRefIndex(), specmodel.SecretsModeOmit)
	assert.Error(t, err)
}

func TestAssembleSettingsIsEmptyForNoSettings(t *testing.T) {
	out, err := assembleSettings(nil, newRefIndex(), specmodel.SecretsModeOmit)
	assert.NoError(t, err)
	assert.Empty(t, out)
}

// omit mode must remove the ciphertext, not merely skip decryption. Leaving it
// would write a value readable only by the installation holding that data key -
// and nothing guarantees a bundle is imported back where it came from.
func TestAssembleSettingsOmitModeClearsCiphertext(t *testing.T) {
	useDataKey(t)
	secret := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "db-password"},
		&entity.Secret{Key: "DB_PASSWORD", Value: entity.NewEncryptedField("hunter2")})

	out, err := assembleSettings([]*entity.Setting{secret}, newRefIndex(), specmodel.SecretsModeOmit)
	assert.NoError(t, err)

	rendered := mustYAML(t, out)
	assert.NotContains(t, rendered, "hpenc", "no ciphertext")
	assert.NotContains(t, rendered, "hunter2", "and certainly no plaintext")
	assert.Contains(t, rendered, "DB_PASSWORD", "the key stays so a reader sees a secret is missing")
}

func TestAssembleSettingsPlaintextModeKeepsTheValueReadable(t *testing.T) {
	useDataKey(t)
	data := &entity.Secret{Key: "DB_PASSWORD", Value: entity.NewEncryptedField("hunter2")}
	secret := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "db-password"}, data)

	// The usecase decrypts before assembly; simulate that here.
	assert.NoError(t, data.Decrypt())

	out, err := assembleSettings([]*entity.Setting{secret}, newRefIndex(), specmodel.SecretsModePlaintext)
	assert.NoError(t, err)
	assert.Contains(t, mustYAML(t, out), "hunter2")
}

// Encrypted mode produces the same readable document plaintext mode does. The
// difference between them is the age envelope around the bundle, applied later
// - not whether the values inside were read.
//
// Leaving the stored ciphertext here instead would defeat the point:
// unwrapping the envelope on another installation would yield a value sealed
// with a data key that installation does not have.
func TestAssembleSettingsEncryptedModeSubstitutesPlaintextToo(t *testing.T) {
	useDataKey(t)
	data := &entity.Secret{Key: "DB_PASSWORD", Value: entity.NewEncryptedField("hunter2")}
	secret := settingWith(t,
		&entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "db-password"}, data)
	assert.NoError(t, data.Decrypt())

	out, err := assembleSettings([]*entity.Setting{secret}, newRefIndex(), specmodel.SecretsModeEncrypted)
	assert.NoError(t, err)

	rendered := mustYAML(t, out)
	assert.Contains(t, rendered, "hunter2", "age protects the bundle; the value inside travels")
	assert.NotContains(t, rendered, "hpenc", "installation-specific ciphertext must not survive")
}
