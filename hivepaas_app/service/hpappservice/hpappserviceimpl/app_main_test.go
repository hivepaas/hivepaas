package hpappserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func setupRoutingTest(t *testing.T) *entity.AppRoutingSettings {
	t.Helper()
	cfg := &config.Config{}
	cfg.HTTPServer.BasePath = "/api/v1"
	cfg.HTTPServer.Port = 8080
	config.SetCurrent(cfg)

	return &entity.AppRoutingSettings{
		Domains: []*entity.AppDomain{{Domain: "hivepaas.example.com"}},
	}
}

func pathOf(domain *entity.AppDomain, path string) *entity.HTTPPathConfig {
	for _, pathCfg := range domain.Paths {
		if pathCfg.Path == path {
			return pathCfg
		}
	}
	return nil
}

func TestSetupRoutingSettingsDefaultAppliesRateLimits(t *testing.T) {
	routing := setupRoutingTest(t)
	(&service{}).SetupRoutingSettingsDefault(routing)
	domain := routing.Domains[0]

	for _, path := range []string{"/api/v1/auth", "/api/v1/system/hivepaas", "/api/v1/cluster"} {
		pathCfg := pathOf(domain, path)
		assert.NotNil(t, pathCfg, "expected a rate limited path config for %s", path)
		assert.Equal(t, base.HTTPPathModePrefix, pathCfg.Mode)
		assert.True(t, pathCfg.RateLimitConfig.Enabled)
		assert.Positive(t, pathCfg.RateLimitConfig.Average)
	}
}

// /system as a whole must stay unlimited: the dashboard polls task status under
// it during a deployment, and a limit there would break the page rather than any
// attacker.
func TestSetupRoutingSettingsDefaultLeavesPolledPathsAlone(t *testing.T) {
	routing := setupRoutingTest(t)
	(&service{}).SetupRoutingSettingsDefault(routing)

	assert.Nil(t, pathOf(routing.Domains[0], "/api/v1/system"),
		"a limit on /system would also catch /system/tasks/:id/status")
}

// It runs on every routing settings update, not only at first setup, so it must
// add to what is there rather than replace it.
func TestSetupRoutingSettingsDefaultKeepsExistingPathConfig(t *testing.T) {
	routing := setupRoutingTest(t)
	domain := routing.Domains[0]
	domain.Paths = []*entity.HTTPPathConfig{{
		Enabled:   true,
		Path:      "/api/v1/auth",
		Mode:      base.HTTPPathModePrefix,
		BasicAuth: &entity.HTTPBasicAuthConfig{Enabled: true, ID: "operator-set-this"},
	}}

	(&service{}).SetupRoutingSettingsDefault(routing)

	authPath := pathOf(domain, "/api/v1/auth")
	assert.NotNil(t, authPath.BasicAuth, "an operator's own settings must survive a resave")
	assert.Equal(t, "operator-set-this", authPath.BasicAuth.ID)
	assert.True(t, authPath.RateLimitConfig.Enabled)
}

// Two saves in a row must not stack up duplicate path entries.
func TestSetupRoutingSettingsDefaultIsIdempotent(t *testing.T) {
	routing := setupRoutingTest(t)
	svc := &service{}

	svc.SetupRoutingSettingsDefault(routing)
	first := len(routing.Domains[0].Paths)
	svc.SetupRoutingSettingsDefault(routing)

	assert.Equal(t, first, len(routing.Domains[0].Paths))
	assert.Equal(t, len(apiRateLimits), first)
}
