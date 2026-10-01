package webhookuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
)

func TestWithheldSecretsNote(t *testing.T) {
	secrets := []*apppreviewservice.WithheldSecret{
		{Name: "API_KEY", EnvVars: []string{"API_KEY", "PAYMENT_URL"}},
		{Name: "STRIPE_KEY", EnvVars: []string{"PAYMENT_URL"}},
	}
	const secretsURL = "https://hp.example.com/projects/prj1/prod/apps/app1/secrets/"

	note := buildWithheldSecretsNote("shop-api", secretsURL, secrets)
	assert.Contains(t, note, "These secrets of `shop-api` are not inheritable")
	assert.Contains(t, note, "> - `API_KEY`, used by `API_KEY`, `PAYMENT_URL`\n")
	assert.Contains(t, note, "> - `STRIPE_KEY`, used by `PAYMENT_URL`\n")
	assert.Contains(t, note, "turn on **Inheritable** for it in the application's [**Secrets**]("+secretsURL+")")

	assert.Contains(t, buildWithheldSecretsNote("shop-api", "", secrets), "application's **Secrets** on",
		"no dashboard address, no link")
	assert.Empty(t, buildWithheldSecretsNote("shop-api", secretsURL, nil), "nothing withheld, no warning")

	comment := buildDeployPreviewComment(true, note)
	assert.Contains(t, comment, note)
}

func TestAppPageURL(t *testing.T) {
	app := &entity.App{ID: "app1", ProjectEnvID: "prj1:prod"}
	assert.Equal(t, "https://hp.example.com/projects/prj1/prod/apps/app1/secrets/",
		appPageURL("https://hp.example.com/", app, "secrets"))
	assert.Empty(t, appPageURL("", app, "secrets"))
}
