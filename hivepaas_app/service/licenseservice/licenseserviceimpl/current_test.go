package licenseserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
)

// No key is Bee wherever it runs, so the swarm is not asked which one it is.
func TestNoKeyIsBeeWithoutAskingTheSwarm(t *testing.T) {
	h := newHarness(t)

	state, err := h.Current(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusNone, state.Status)
	assert.Equal(t, base.LicensePlanBee, state.Plan)
	assert.Equal(t, license.BeeLimits, state.Limits)
	assert.Zero(t, h.docker.clusterCalls)
	assert.Zero(t, h.trustCalls)
}

func TestTheStoredKeyIsInForce(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload())

	state, err := h.Current(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
	assert.Equal(t, base.LicensePlanHive, state.Plan)
	assert.Equal(t, license.Limits{base.LicenseLimitNodes: 10, base.LicenseLimitUsers: 10}, state.Limits)
	assert.Equal(t, thisCluster, state.InstallationID)
}

// A key that names another swarm is not in force here.
func TestAKeyForAnotherInstallationIsBee(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload(func(p *license.Payload) { p.Installations = []string{otherCluster} }))

	state, err := h.Current(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusOtherInstallation, state.Status)
	assert.Equal(t, base.LicensePlanBee, state.Plan)
}

// Without the swarm's ID there is no telling whether the key is this
// installation's. Reading that as Bee would refuse a paying customer's
// creations as over the free plan's limits; the error is passed on instead.
func TestNoSwarmIDIsAnErrorNotBee(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload())
	h.docker.clusterID, h.docker.clusterErr = "", hperrors.NewNotFound("Swarm cluster")

	state, err := h.Current(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	assert.Nil(t, state)
}

// The signatures of a key are verified again when the key changes, or when
// whether development keys count does, and not on every read.
func TestTheKeyIsVerifiedAgainOnlyWhenItChanges(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.storeKey(t, payload())

	for range 3 {
		_, err := h.Current(ctx)
		assert.NoError(t, err)
	}
	assert.Equal(t, 1, h.trustCalls)
	assert.Equal(t, 3, h.settings.reads, "the setting is read every time: another process may have changed it")

	h.storeKey(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanColony }))
	state, err := h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicensePlanColony, state.Plan)
	assert.Equal(t, 2, h.trustCalls)

	h.devEnv = true
	_, err = h.Current(ctx)
	assert.NoError(t, err)
	_, err = h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, h.trustCalls)
}

// A development key is in force only in development.
func TestADevelopmentKeyIsInForceOnlyInDevelopment(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.settings.store(t, h.dev.keyText(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanColony })))

	state, err := h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusInvalid, state.Status)
	assert.Equal(t, base.LicensePlanBee, state.Plan)
	assert.Contains(t, state.Reason, "development")

	h.devEnv = true
	state, err = h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
	assert.Equal(t, base.LicensePlanColony, state.Plan)
	assert.True(t, state.Development)
}

// Compiled-in keys that do not parse are a build mistake. They make the stored
// key invalid, which is Bee, rather than every creation an error.
func TestKeysThatDoNotParseLeaveTheKeyInvalid(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload())
	h.trustErr = hperrors.Wrap(hperrors.ErrLicenseSigningKeyInvalid)

	state, err := h.Current(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusInvalid, state.Status)
	assert.Equal(t, base.LicensePlanBee, state.Plan)
}

// The latest time seen is written as the clock moves on, at most once an hour,
// and never lowered.
func TestTheLatestTimeSeenIsWrittenHourly(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.status.seenAt = testNow.Add(-2 * time.Hour)

	current := func(at time.Time) {
		t.Helper()
		h.now = at
		_, err := h.Current(ctx)
		assert.NoError(t, err)
	}
	current(testNow)
	current(testNow.Add(30 * time.Minute))
	current(testNow.Add(59 * time.Minute))
	current(testNow.Add(time.Hour))
	current(testNow.Add(time.Hour + time.Minute))
	current(testNow.Add(-3 * day))

	assert.Equal(t, []time.Time{testNow, testNow.Add(time.Hour)}, h.status.recorded)
	assert.Equal(t, 1, h.status.gets, "read once, then kept")
}

// A clock set back more than a day is read as the latest time seen: an expired
// key does not come back into force.
func TestAClockSetBackIsReadAsTheLatestTimeSeen(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	key := payload(func(p *license.Payload) { p.GraceDays = 0 })
	h.storeKey(t, key)
	h.status.seenAt = key.ExpiresAt.Add(2 * day)

	h.now = key.ExpiresAt.Add(-5 * day)
	state, err := h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusExpired, state.Status)
	assert.Equal(t, base.LicensePlanBee, state.Plan)
	assert.Empty(t, h.status.recorded, "a clock set back is never written")

	// Corrections of minutes or hours are an honest clock's, and believed.
	h.now = key.ExpiresAt.Add(-time.Hour)
	h.status.seenAt = key.ExpiresAt.Add(time.Hour)
	h.service.clock = clockState{}
	state, err = h.Current(ctx)
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
}

func TestTheInstallationIDIsTheSwarms(t *testing.T) {
	h := newHarness(t)

	id, err := h.InstallationID(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, thisCluster, id)
}
