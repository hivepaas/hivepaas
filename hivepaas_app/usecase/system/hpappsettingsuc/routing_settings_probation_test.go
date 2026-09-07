package hpappsettingsuc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestResolveProbationWindow(t *testing.T) {
	t.Run("absent means the default", func(t *testing.T) {
		assert.Equal(t, probationWindowDefault, resolveProbationWindow(0))
	})

	// There is no way to ask for no trial at all. A caller that cannot confirm is
	// the caller this exists for, so a script must not be able to opt itself out
	// by sending a zero or a negative window.
	t.Run("a negative window cannot turn the trial off", func(t *testing.T) {
		assert.Equal(t, probationWindowDefault, resolveProbationWindow(-time.Hour))
	})

	t.Run("clamped at both ends", func(t *testing.T) {
		assert.Equal(t, probationWindowMin, resolveProbationWindow(time.Second))
		assert.Equal(t, probationWindowMax, resolveProbationWindow(24*time.Hour))
	})

	t.Run("a window in range is kept", func(t *testing.T) {
		assert.Equal(t, 5*time.Minute, resolveProbationWindow(5*time.Minute))
	})

	// Confirmation is refused until the settle delay has passed, so any window at
	// or near it hands the operator a countdown they cannot answer - and they lock
	// themselves out using the mechanism that exists to stop exactly that. The
	// smallest window has to leave a usable margin on the far side of it.
	t.Run("every allowed window leaves time to answer", func(t *testing.T) {
		usableAtMin := probationWindowMin - entity.SettingsProbationSettleDelay
		assert.Greater(t, usableAtMin, 30*time.Second,
			"the smallest window must leave a margin a person can act inside")

		usableAtDefault := probationWindowDefault - entity.SettingsProbationSettleDelay
		assert.Greater(t, usableAtDefault, 2*time.Minute,
			"the default must leave time to go and check the change actually works")
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
	snapshot := entity.SnapshotOf(setting)

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
