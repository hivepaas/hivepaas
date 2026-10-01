package apppreviewserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apppreviewservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func TestWithheldSecretsAreThoseNotInheritableTheVariablesUse(t *testing.T) {
	apiKey := &entity.Setting{ID: "s1", Name: "API_KEY"}
	stripe := &entity.Setting{ID: "s2", Name: "STRIPE_KEY"}
	shared := &entity.Setting{ID: "s3", Name: "SENTRY_DSN", Inheritable: true}
	envVar := func(key string, secrets ...*entity.Setting) *envvarservice.EnvVar {
		v := &envvarservice.EnvVar{EnvVar: &entity.EnvVar{Key: key}}
		for _, s := range secrets {
			v.AddRefSecretSetting(s)
		}
		return v
	}

	got := withheldSecrets([]*envvarservice.EnvVar{
		envVar("PAYMENT_URL", stripe, apiKey),
		envVar("API_KEY", apiKey),
		envVar("SENTRY_DSN", shared),
		envVar("PORT"),
		// A build variable of the same name as a runtime one is named once.
		envVar("API_KEY", apiKey),
	})

	assert.Equal(t, []*apppreviewservice.WithheldSecret{
		{Name: "API_KEY", EnvVars: []string{"API_KEY", "PAYMENT_URL"}},
		{Name: "STRIPE_KEY", EnvVars: []string{"PAYMENT_URL"}},
	}, got)
}

func TestNoWithheldSecretsWhenAllAreInheritable(t *testing.T) {
	v := &envvarservice.EnvVar{EnvVar: &entity.EnvVar{Key: "DSN"}}
	v.AddRefSecretSetting(&entity.Setting{ID: "s1", Name: "DSN", Inheritable: true})
	assert.Empty(t, withheldSecrets([]*envvarservice.EnvVar{v}))
}
