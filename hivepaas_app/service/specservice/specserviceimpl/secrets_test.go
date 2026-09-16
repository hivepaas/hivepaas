package specserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

func storedSecret(t *testing.T, plain string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "s1", Type: base.SettingTypeSecret, Name: "db-password"}
	assert.NoError(t, setting.SetData(&entity.Secret{
		Key: "DB_PASSWORD", Value: entity.NewEncryptedField(plain),
	}))
	// Return it the way a row read back from the database looks: no parse cache.
	return &entity.Setting{
		ID: setting.ID, Type: setting.Type, Name: setting.Name, Data: setting.Data,
	}
}

func TestRevealSettingSecretsDoesNothingInOmitMode(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModeOmit))

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	assert.True(t, data.Value.IsEncrypted(), "omit decrypts nothing")
}

func TestRevealSettingSecretsDecryptsInBothSecretModes(t *testing.T) {
	for _, mode := range []specmodel.SecretsMode{
		specmodel.SecretsModePlaintext,
		specmodel.SecretsModeEncrypted,
	} {
		useDataKey(t)
		setting := storedSecret(t, "hunter2")

		assert.NoError(t, revealSettingSecrets(setting, mode))

		data, err := setting.AsSecret()
		assert.NoError(t, err)
		plain, err := data.Value.GetPlain()
		assert.NoError(t, err)
		assert.Equal(t, "hunter2", plain, "mode %v", mode)
	}
}

// A type holding no secrets must not error just because a mode asks for them.
func TestRevealSettingSecretsIgnoresTypesWithoutSecrets(t *testing.T) {
	useDataKey(t)
	setting := &entity.Setting{ID: "s2", Type: base.SettingTypeScript}
	assert.NoError(t, setting.SetData(&entity.Script{Data: "echo hi"}))

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModePlaintext))
}

// A setting inherited from an outer scope is read through, not owned. Its
// secrets belong to whoever defined it, which is what stops a project-scope
// export from extracting global ones.
func TestRevealSettingSecretsSkipsInheritedSettings(t *testing.T) {
	useDataKey(t)
	setting := storedSecret(t, "hunter2")
	setting.ObjectID = "global_owner"
	setting.CurrentObjectID = "some_project"

	assert.NoError(t, revealSettingSecrets(setting, specmodel.SecretsModePlaintext))

	data, err := setting.AsSecret()
	assert.NoError(t, err)
	assert.True(t, data.Value.IsEncrypted(), "an inherited secret is not this scope's to reveal")
}
