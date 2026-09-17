package entity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func newTestAppTemplateSettings() *AppTemplateSettings {
	return &AppTemplateSettings{
		Source:        "official",
		Template:      "postgres",
		Title:         "PostgreSQL",
		Version:       "17",
		Variant:       "alpine",
		ImageOverride: "postgres:18.7-alpine3.24",
		Params: map[string]*AppTemplateParam{
			"username": {Value: "app"},
			"password": {Secret: NewEncryptedField("generated-password")},
		},
		Base: AppTemplateBase{
			Revision:       "0123456789abcdef0123456789abcdef01234567",
			Release:        "17.6",
			TemplateSHA256: "aa",
			Rendered:       "deployment: {}\n",
			RenderedSHA256: "bb",
			AppliedAt:      time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC),
		},
	}
}

func TestAppTemplateSettingsRoundTrip(t *testing.T) {
	useDataKey(t)
	setting := &Setting{Type: base.SettingTypeAppTemplate}

	assert.NoError(t, setting.SetData(newTestAppTemplateSettings()))
	assert.NotContains(t, setting.Data, "generated-password", "a secret parameter is stored encrypted")

	restored := &Setting{Type: base.SettingTypeAppTemplate, Data: setting.Data}
	parsed, err := restored.AsAppTemplateSettings()
	assert.NoError(t, err)
	assert.Equal(t, "app", parsed.Params["username"].Value)
	assert.NoError(t, parsed.Decrypt())
	password, err := parsed.Params["password"].Secret.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "generated-password", password)
	assert.Equal(t, "17.6", parsed.Base.Release)
	assert.Equal(t, "postgres:18.7-alpine3.24", parsed.ImageOverride)
}

func TestAppTemplateSettingsOmitsSecrets(t *testing.T) {
	useDataKey(t)
	settings := newTestAppTemplateSettings()

	count, err := OmitSecrets(settings)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.True(t, settings.Params["password"].Secret.IsEmpty())
	encoded, err := json.Marshal(settings)
	assert.NoError(t, err)
	assert.NotContains(t, string(encoded), `"secret"`, "an empty secret is left out")
}
