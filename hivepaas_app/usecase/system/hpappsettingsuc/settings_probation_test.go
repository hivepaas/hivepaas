package hpappsettingsuc

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestResolveProbationWindow(t *testing.T) {
	routing := base.SettingTypeAppRouting
	service := base.SettingTypeHivePaaSService

	t.Run("absent means the default", func(t *testing.T) {
		assert.Equal(t, probationWindowDefault, resolveProbationWindow(0, routing))
	})

	// There is no way to ask for no trial at all. A caller that cannot confirm is
	// the caller this exists for, so a script must not be able to opt itself out
	// by sending a zero or a negative window.
	t.Run("a negative window cannot turn the trial off", func(t *testing.T) {
		assert.Equal(t, probationWindowDefault, resolveProbationWindow(-time.Hour, routing))
	})

	t.Run("clamped at the top", func(t *testing.T) {
		assert.Equal(t, probationWindowMax, resolveProbationWindow(24*time.Hour, routing))
	})

	t.Run("a window in range is kept", func(t *testing.T) {
		assert.Equal(t, 5*time.Minute, resolveProbationWindow(5*time.Minute, routing))
	})

	// The floor moves with the setting type. A change that restarts traefik cannot
	// be confirmed for over a minute, so the window it is given has to outlast
	// that - a caller asking for the routing minimum here would otherwise get a
	// countdown that runs out before the confirm button ever works.
	t.Run("the floor covers the settle delay of each type", func(t *testing.T) {
		for _, typ := range []base.SettingType{routing, service} {
			usable := resolveProbationWindow(time.Second, typ) - entity.SettleDelayFor(typ)
			assert.GreaterOrEqual(t, usable, probationAnswerMargin,
				"the smallest window for %s must leave time to answer", typ)
		}
		assert.Greater(t, resolveProbationWindow(time.Second, service),
			resolveProbationWindow(time.Second, routing))
	})
}

func TestConfirmableFrom(t *testing.T) {
	appliedAt := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	args := &entity.TaskSettingsRevertArgs{AppliedAt: appliedAt}

	// A confirmation arriving before traefik has picked the new labels up would
	// have traveled through the configuration being replaced, so it vouches for
	// a router that is not live yet.
	assert.Equal(t, appliedAt.Add(entity.SettingsProbationSettleDelay), args.ConfirmableFrom())
	assert.Greater(t, entity.SettingsProbationSettleDelay, 15*time.Second,
		"the settle delay has to outlast traefik's swarm provider poll interval")
}

func TestSettingSnapshotRoundTrip(t *testing.T) {
	setting := &entity.Setting{
		Type:    base.SettingTypeAppRouting,
		Data:    `{"exposePublicly":true}`,
		Version: 3,
	}
	snapshot := entity.SettingSnapshotOf(setting)

	// Move the setting on the way an update does, and parse it - which is what
	// leaves the new payload cached on the setting.
	setting.Data = `{"exposePublicly":false}`
	setting.Version = 4
	changed, err := setting.AsAppRoutingSettings()
	assert.NoError(t, err)
	assert.False(t, changed.ExposePublicly)

	snapshot.RestoreTo(setting)
	assert.Equal(t, `{"exposePublicly":true}`, setting.Data)
	assert.Equal(t, 3, setting.Version)

	// The cached parse has to go with it. Left in place, every reader that goes
	// through MustAsX() - the apply path included - would keep working from the
	// settings that were just reverted away from, while the column says otherwise.
	restored, err := setting.AsAppRoutingSettings()
	assert.NoError(t, err)
	assert.True(t, restored.ExposePublicly)
}

// A confirmation is only worth something once the change it vouches for is the
// one being served. Routing settings need traefik to notice new labels; service
// settings take traefik down and bring it back, which is a different order of
// wait entirely, and treating them the same would let somebody confirm a proxy
// change while the old traefik was still answering.
func TestConfirmableFromDependsOnWhatRestarts(t *testing.T) {
	appliedAt := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)

	routing := &entity.TaskSettingsRevertArgs{AppliedAt: appliedAt, SettingType: base.SettingTypeAppRouting}
	service := &entity.TaskSettingsRevertArgs{AppliedAt: appliedAt, SettingType: base.SettingTypeHivePaaSService}

	assert.Equal(t, appliedAt.Add(entity.SettingsProbationSettleDelay), routing.ConfirmableFrom())
	assert.Equal(t, appliedAt.Add(entity.SettingsProbationRestartSettleDelay), service.ConfirmableFrom())

	assert.Greater(t, service.ConfirmableFrom().Sub(routing.ConfirmableFrom()), time.Duration(0),
		"a change that restarts traefik has to wait longer than one that only relabels")
	// It has to outlast the start_period of everything the change can restart:
	// traefik at 60s, and the main app at 120s when the same request also carries
	// a replica or worker setting.
	assert.GreaterOrEqual(t, entity.SettingsProbationRestartSettleDelay, 120*time.Second)

	// Every window still has to leave room to answer on the slower of the two.
	assert.GreaterOrEqual(t, resolveProbationWindow(0, base.SettingTypeHivePaaSService)-
		entity.SettingsProbationRestartSettleDelay, probationAnswerMargin)
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

// The window the dashboard asks for has to leave real time to answer once the
// settle delay is spent, or the operator watches a countdown they cannot act on.
// The dashboard sends 5m; this is what that buys after the wait.
func TestDashboardWindowLeavesTimeToAnswer(t *testing.T) {
	const dashboardWindow = 5 * time.Minute

	usable := resolveProbationWindow(dashboardWindow, base.SettingTypeHivePaaSService) -
		entity.SettingsProbationRestartSettleDelay
	assert.GreaterOrEqual(t, usable, 2*time.Minute,
		"a proxy change leaves too little time to check anything and confirm")
}
