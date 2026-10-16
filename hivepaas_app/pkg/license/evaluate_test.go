package license

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func TestEvaluate(t *testing.T) {
	prod, dev := newPair(t, "lic-2026"), newPair(t, "dev")
	trust := &Trust{Prod: publicKeysOf(t, prod...), Dev: publicKeysOf(t, dev...), AllowDev: true}
	hiveLimits := Limits{base.LicenseLimitNodes: 10, base.LicenseLimitUsers: 10}

	// key is the stored key's text for a payload, signed by the production keys
	// unless it says otherwise.
	key := func(change func(p *Payload), keys ...*testKey) string {
		p := validPayload()
		if change != nil {
			change(p)
		}
		if len(keys) == 0 {
			keys = prod
		}
		return token(seal(t, marshal(t, p), Context, keys...))
	}
	expiresAt := validPayload().ExpiresAt
	graceEndsAt := expiresAt.Add(14 * day)

	cases := []struct {
		name         string
		keyText      string
		installation string
		now          time.Time
		movedAt      time.Time

		status      base.LicenseStatus
		plan        base.LicensePlan
		limits      Limits
		graceEndsAt time.Time
		reason      string
		development bool
	}{
		{name: "no key", status: base.LicenseStatusNone, plan: base.LicensePlanBee, limits: BeeLimits},
		{name: "a key that does not verify", keyText: "not a key!",
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nneither base64 nor JSON"},
		{name: "version 2", keyText: key(func(p *Payload) { p.Version = 2 }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\npayload version 2 is not one this HivePaaS reads"},
		{name: "plan bee", keyText: key(func(p *Payload) { p.Plan = base.LicensePlanBee }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nplan \"bee\" is never in a key"},
		{name: "an unknown plan", keyText: key(func(p *Payload) { p.Plan = "swarm" }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nplan \"swarm\" is not one this HivePaaS knows"},
		{name: "no installations", keyText: key(func(p *Payload) { p.Installations = nil }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nno installation"},
		{name: "expires when issued", keyText: key(func(p *Payload) { p.ExpiresAt = p.IssuedAt }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nexpiresAt is not after issuedAt"},
		{name: "a negative limit", keyText: key(func(p *Payload) { p.Limits["nodes"] = -1 }),
			status: base.LicenseStatusInvalid, plan: base.LicensePlanBee, limits: BeeLimits,
			reason: "This license key could not be verified\n\nlimit \"nodes\" is negative"},
		{name: "another installation's", keyText: key(nil), installation: "cluster-b",
			status: base.LicenseStatusOtherInstallation, plan: base.LicensePlanBee, limits: BeeLimits},
		{name: "valid", keyText: key(nil),
			status: base.LicenseStatusValid, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "valid, signed by the development keys", keyText: key(nil, dev...), development: true,
			status: base.LicenseStatusValid, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "valid, an instant before it expires", keyText: key(nil), now: expiresAt.Add(-time.Nanosecond),
			status: base.LicenseStatusValid, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "Apiary, one of several installations", installation: "cluster-c",
			keyText: key(func(p *Payload) {
				p.Plan = base.LicensePlanApiary
				p.Installations = []string{"cluster-b", "cluster-c", "cluster-d"}
				p.Limits = map[string]int{"nodes": 100, "users": 250, "previews": 50}
			}),
			status: base.LicenseStatusValid, plan: base.LicensePlanApiary, graceEndsAt: graceEndsAt,
			limits: Limits{base.LicenseLimitNodes: 100, base.LicenseLimitUsers: 250, base.LicenseLimitPreviews: 50}},
		{name: "trial", keyText: key(func(p *Payload) { p.Trial, p.GraceDays = true, 0 }),
			status: base.LicenseStatusTrial, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: expiresAt},
		{name: "trial, run out", keyText: key(func(p *Payload) { p.Trial, p.GraceDays = true, 0 }), now: expiresAt,
			status: base.LicenseStatusExpired, plan: base.LicensePlanBee, limits: BeeLimits, graceEndsAt: expiresAt},
		{name: "grace, from the instant it expires", keyText: key(nil), now: expiresAt,
			status: base.LicenseStatusGrace, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "grace, an instant before it ends", keyText: key(nil), now: graceEndsAt.Add(-time.Nanosecond),
			status: base.LicenseStatusGrace, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "expired, from the instant grace ends", keyText: key(nil), now: graceEndsAt,
			status: base.LicenseStatusExpired, plan: base.LicensePlanBee, limits: BeeLimits, graceEndsAt: graceEndsAt},
		{name: "moved, within grace of the move", keyText: key(nil), movedAt: testNow.Add(-13 * day),
			status: base.LicenseStatusMoved, plan: base.LicensePlanHive, limits: hiveLimits,
			graceEndsAt: testNow.Add(day)},
		{name: "moved, past grace of the move", keyText: key(nil), movedAt: testNow.Add(-14 * day),
			status: base.LicenseStatusMoved, plan: base.LicensePlanBee, limits: BeeLimits, graceEndsAt: testNow},
		{name: "moved, the key's own grace ending sooner", keyText: key(nil), now: graceEndsAt.Add(-day),
			movedAt: graceEndsAt.Add(-2 * day),
			status:  base.LicenseStatusMoved, plan: base.LicensePlanHive, limits: hiveLimits, graceEndsAt: graceEndsAt},
		{name: "moved, a trial: no grace", keyText: key(func(p *Payload) { p.Trial, p.GraceDays = true, 0 }),
			movedAt: testNow.Add(-time.Minute),
			status:  base.LicenseStatusMoved, plan: base.LicensePlanBee, limits: BeeLimits,
			graceEndsAt: testNow.Add(-time.Minute)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &EvalInput{KeyText: tc.keyText, InstallationID: tc.installation, Now: tc.now, MovedAt: tc.movedAt}
			if in.InstallationID == "" {
				in.InstallationID = testCluster
			}
			if in.Now.IsZero() {
				in.Now = testNow
			}
			if in.KeyText != "" {
				in.Verified, in.VerifyErr = Verify(trust, in.KeyText)
			}

			state := Evaluate(in)
			assert.Equal(t, tc.status, state.Status)
			assert.Equal(t, tc.plan, state.Plan)
			assert.Equal(t, tc.limits, state.Limits)
			assert.Equal(t, tc.graceEndsAt, state.GraceEndsAt)
			assert.Equal(t, tc.reason, state.Reason)
			assert.Equal(t, tc.development, state.Development)
			assert.Equal(t, in.InstallationID, state.InstallationID)
			assert.Equal(t, tc.movedAt, state.MovedAt)
			switch tc.status {
			case base.LicenseStatusNone, base.LicenseStatusInvalid:
				assert.Nil(t, state.Key, "an unverified payload is never read")
			case base.LicenseStatusValid, base.LicenseStatusTrial, base.LicenseStatusGrace, base.LicenseStatusExpired,
				base.LicenseStatusOtherInstallation, base.LicenseStatusMoved:
				assert.NotNil(t, state.Key)
			}
		})
	}
}

// Each State's limits are its own: a caller changing them changes neither Bee's
// nor another State's.
func TestEvaluateLimitsAreTheStatesOwn(t *testing.T) {
	state := Evaluate(&EvalInput{InstallationID: testCluster, Now: testNow})
	state.Limits[base.LicenseLimitApps] = 1000

	assert.Equal(t, 20, BeeLimits[base.LicenseLimitApps])
	assert.Equal(t, 20, Evaluate(&EvalInput{InstallationID: testCluster, Now: testNow}).Limits[base.LicenseLimitApps])
}
