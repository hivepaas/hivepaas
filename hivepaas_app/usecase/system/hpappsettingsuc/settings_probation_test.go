package hpappsettingsuc

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// The sweep re-applies the proxy topology to every app's labels, and only the
// HivePaaS service settings carry one. Running it for any other kind of trial
// would recreate every app's labels for a change that is not in them - so the
// guard is what keeps confirming a routing tweak from touching the whole cluster.
//
// A zero UC and an empty transaction are enough: the guard has to return before
// it reaches either, and a nil taskRepo is what proves it did.
func TestAppLabelsSweepOnlyForServiceSettings(t *testing.T) {
	uc := &UC{}

	for _, typ := range []base.SettingType{base.SettingTypeAppRouting, base.SettingTypeTraefikConfig} {
		t.Run(string(typ), func(t *testing.T) {
			tasks, err := uc.appLabelsSweepOnConfirm(context.Background(), database.Tx{},
				&entity.TaskSettingsRevertArgs{SettingType: typ, AppID: "app-1"})

			assert.NoError(t, err)
			assert.Empty(t, tasks, "confirming a %s change must not sweep every app's labels", typ)
		})
	}
}

// Equal is what decides whether a service settings change goes on trial at all.
// A proxy field it does not look at is a proxy field that applies with no way back.
func TestProxySettingsEqualCoversEveryField(t *testing.T) {
	settings := entity.HivePaaSProxySettings{
		ProxyProvider: "cloudflare",
		TrustedIPs:    []string{"10.0.0.0/8"},
		ProxyHops:     2,
	}

	assert.True(t, settings.Equal(&entity.HivePaaSProxySettings{
		ProxyProvider: "cloudflare",
		TrustedIPs:    []string{"10.0.0.0/8"},
		ProxyHops:     2,
	}))

	differs := map[string]entity.HivePaaSProxySettings{
		"provider":   {ProxyProvider: "other", TrustedIPs: []string{"10.0.0.0/8"}, ProxyHops: 2},
		"trustedIPs": {ProxyProvider: "cloudflare", TrustedIPs: []string{"10.0.0.0/9"}, ProxyHops: 2},
		"extra IP":   {ProxyProvider: "cloudflare", TrustedIPs: []string{"10.0.0.0/8", "1.1.1.1"}, ProxyHops: 2},
		"no IPs":     {ProxyProvider: "cloudflare", TrustedIPs: nil, ProxyHops: 2},
		"proxyHops":  {ProxyProvider: "cloudflare", TrustedIPs: []string{"10.0.0.0/8"}, ProxyHops: 3},
	}
	for name, other := range differs {
		t.Run(name, func(t *testing.T) {
			assert.False(t, settings.Equal(&other), "a change to %s must put the settings on trial", name)
		})
	}

	// The struct must not grow a field this forgets about.
	assert.Equal(t, 3, reflect.TypeOf(entity.HivePaaSProxySettings{}).NumField(),
		"HivePaaSProxySettings gained a field - teach Equal about it, or a change to it applies with no way back")
}

// A routing change rewrites swarm service labels, which does not recreate a task
// - so there is no failed update for swarm to roll back, and nothing to check.
// Checking anyway would mean reading traefik's trusted IPs on a change that never
// touched them, and refusing confirmations over a mismatch that means nothing.
//
// A zero UC is enough: it has to return before it reaches any dependency.
func TestProxyLivenessCheckOnlyForServiceSettings(t *testing.T) {
	uc := &UC{}

	err := uc.ensureProxySettingsAreLive(context.Background(), database.Tx{},
		&entity.TaskSettingsRevertArgs{SettingType: base.SettingTypeAppRouting})

	assert.NoError(t, err)
}
