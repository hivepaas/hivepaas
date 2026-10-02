package githubappuc

import (
	"testing"

	gogithub "github.com/google/go-github/v85/github"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The app's hook on GitHub is given the URL and the secret HivePaaS checks its
// deliveries with: a delivery signed with any other secret is refused.
func TestTheAppsHookCarriesTheSecretDeliveriesAreCheckedWith(t *testing.T) {
	app := &entity.GithubApp{
		WebhookURL:    "https://hp.example.com/api/v1/webhooks/S1",
		WebhookSecret: entity.NewEncryptedField("s3cr3t"),
	}

	set, err := appHookConfig(app)
	assert.NoError(t, err)
	cfg := &gogithub.HookConfig{}
	set(cfg)

	assert.Equal(t, "https://hp.example.com/api/v1/webhooks/S1", cfg.GetURL())
	assert.Equal(t, "json", cfg.GetContentType())
	assert.Equal(t, "s3cr3t", cfg.GetSecret())
}
