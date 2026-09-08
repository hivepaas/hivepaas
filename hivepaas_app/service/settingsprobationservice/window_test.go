package settingsprobationservice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestResolveWindow(t *testing.T) {
	settle := entity.SettingsProbationSettleDelay
	restart := entity.SettingsProbationAppRestartSettleDelay

	t.Run("absent means the default", func(t *testing.T) {
		assert.Equal(t, WindowDefault, ResolveWindow(0, settle))
	})

	// There is no way to ask for no trial at all. A caller that cannot confirm is
	// the caller this exists for, so a script must not be able to opt itself out
	// by sending a zero or a negative window.
	t.Run("a negative window cannot turn the trial off", func(t *testing.T) {
		assert.Equal(t, WindowDefault, ResolveWindow(-time.Hour, settle))
	})

	t.Run("clamped at the top", func(t *testing.T) {
		assert.Equal(t, WindowMax, ResolveWindow(24*time.Hour, settle))
	})

	t.Run("a window in range is kept", func(t *testing.T) {
		assert.Equal(t, 5*time.Minute, ResolveWindow(5*time.Minute, settle))
	})

	// The floor moves with the settle delay. A change that takes the app down
	// cannot be confirmed for two minutes, so the window it is given has to
	// outlast that - a caller asking for the ordinary minimum here would otherwise
	// get a countdown that runs out before the confirm button ever works.
	t.Run("the floor leaves time to answer at every settle delay", func(t *testing.T) {
		for _, delay := range []time.Duration{settle, restart} {
			usable := ResolveWindow(time.Second, delay) - delay
			assert.GreaterOrEqual(t, usable, answerMargin,
				"the smallest window for a %v settle must leave time to answer", delay)
		}
		assert.Greater(t, ResolveWindow(time.Second, restart), ResolveWindow(time.Second, settle))
	})

	// A caller passing nothing, or less than the shared bound, still gets it. The
	// settle delay is a correctness bound before it is a preference.
	t.Run("the shared floor cannot be undercut", func(t *testing.T) {
		assert.Equal(t, ResolveWindow(time.Second, settle), ResolveWindow(time.Second, 0))
		assert.Equal(t, ResolveWindow(time.Second, settle), ResolveWindow(time.Second, time.Second))
	})
}

// The settle delay guards one thing: a confirmation that the previous
// configuration served. Two things decide when that stops being possible, and
// the delay has to outlast the slower.
//
// Measured on a single-node swarm, three runs agreeing to within 50ms: replacing
// traefik's task takes the old one out of service at 3.2s and has the new one
// serving at 4.2s. Traefik's swarm provider poll is 15s, and is the slower of the
// two - so a label change, which restarts nothing, is what sets the number.
func TestSettleDelayOutlastsWhatItGuards(t *testing.T) {
	const (
		traefikSwarmPollInterval = 15 * time.Second
		measuredTaskReplacement  = 5 * time.Second
	)

	assert.Greater(t, entity.SettingsProbationSettleDelay, traefikSwarmPollInterval,
		"a label change is not live until traefik polls; confirming before that vouches for the old routers")
	assert.Greater(t, entity.SettingsProbationSettleDelay, measuredTaskReplacement,
		"a traefik command change is not live until its task is replaced")
}

func TestConfirmableFrom(t *testing.T) {
	appliedAt := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)

	// Read from the record: what a change disturbs is known when it is applied.
	args := &entity.TaskSettingsRevertArgs{
		AppliedAt:     appliedAt,
		ConfirmableAt: appliedAt.Add(entity.SettingsProbationAppRestartSettleDelay),
	}
	assert.Equal(t, appliedAt.Add(entity.SettingsProbationAppRestartSettleDelay), args.ConfirmableFrom())

	// A task written before the field existed carries the zero time, and must not
	// become confirmable the instant it is armed.
	legacy := &entity.TaskSettingsRevertArgs{AppliedAt: appliedAt}
	assert.Equal(t, appliedAt.Add(entity.SettingsProbationSettleDelay), legacy.ConfirmableFrom())
	assert.True(t, legacy.ConfirmableFrom().After(appliedAt))
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

// The window the dashboard asks for has to leave real time to answer once the
// settle delay is spent, or the operator watches a countdown they cannot act on.
// The dashboard sends 5m; this is what that buys after the wait.
func TestDashboardWindowLeavesTimeToAnswer(t *testing.T) {
	const dashboardWindow = 5 * time.Minute

	for _, delay := range []time.Duration{
		entity.SettingsProbationSettleDelay,
		entity.SettingsProbationAppRestartSettleDelay,
	} {
		usable := ResolveWindow(dashboardWindow, delay) - delay
		assert.GreaterOrEqual(t, usable, 2*time.Minute,
			"a change with a %v settle leaves too little time to check anything and confirm", delay)
	}
}
