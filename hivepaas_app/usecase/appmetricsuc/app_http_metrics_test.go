package appmetricsuc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func appWithRouting(routing *entity.AppRoutingSettings) *entity.App {
	app := &entity.App{ID: "a1"}
	if routing != nil {
		setting := &entity.Setting{Type: base.SettingTypeAppRouting}
		setting.MustSetData(routing)
		app.Settings = []*entity.Setting{setting}
	}
	return app
}

// Only an app reached by an enabled domain has requests through the proxy.
func TestIsExposed(t *testing.T) {
	domain := func(enabled bool) []*entity.AppDomain {
		return []*entity.AppDomain{{Domain: "a.test", Enabled: enabled}}
	}
	assert.True(t, isExposed(appWithRouting(&entity.AppRoutingSettings{ExposePublicly: true, Domains: domain(true)})))
	assert.False(t, isExposed(appWithRouting(&entity.AppRoutingSettings{ExposePublicly: true, Domains: domain(false)})))
	assert.False(t, isExposed(appWithRouting(&entity.AppRoutingSettings{ExposePublicly: false, Domains: domain(true)})))
	assert.False(t, isExposed(appWithRouting(nil)))
}
