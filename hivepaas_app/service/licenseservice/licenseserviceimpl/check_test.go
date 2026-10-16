package licenseserviceimpl

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// tx stands for the creating transaction. The fakes never use it; Check only
// asks that it is one.
var tx = database.Tx{}

// A kind the plan does not limit is neither locked nor counted, so a paid plan
// never waits on another creation.
func TestAnUnlimitedKindIsNotLockedOrCounted(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload())

	for _, limit := range []base.LicenseLimit{base.LicenseLimitApps, base.LicenseLimitFunctions,
		base.LicenseLimitPreviews, base.LicenseLimitSignInProviders} {
		assert.NoError(t, h.Check(context.Background(), tx, licenseservice.Adding(limit, 1000)), limit)
	}
	assert.Empty(t, h.usage.asked)
	assert.Zero(t, h.docker.nodeCalls)
}

// Bee allows 20 apps: the count is taken under the lock, and a creation that
// would pass the limit is refused - also on an installation already over it,
// such as a beta install, which keeps all it has.
func TestACreationPastTheLimitIsRefused(t *testing.T) {
	for _, tc := range []struct {
		count, adding int
		refused       bool
	}{
		{count: 0, adding: 1},
		{count: 18, adding: 2},
		{count: 19, adding: 1},
		{count: 20, adding: 1, refused: true},
		{count: 19, adding: 2, refused: true},
		{count: 25, adding: 1, refused: true},
	} {
		t.Run(fmt.Sprintf("%d+%d", tc.count, tc.adding), func(t *testing.T) {
			h := newHarness(t)
			h.usage.counts[base.LicenseLimitApps] = tc.count

			err := h.Check(context.Background(), tx, licenseservice.Adding(base.LicenseLimitApps, tc.adding))

			if tc.refused {
				assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, []string{"lock apps", "count apps"}, h.usage.asked)
		})
	}
}

// Each limit is checked against its own count: functions are not apps.
func TestEachKindHasItsOwnCount(t *testing.T) {
	h := newHarness(t)
	h.usage.counts = map[base.LicenseLimit]int{base.LicenseLimitApps: 20, base.LicenseLimitFunctions: 3,
		base.LicenseLimitPreviews: 2, base.LicenseLimitUsers: 1}
	ctx := context.Background()

	assert.NoError(t, h.Check(ctx, tx, licenseservice.Adding(base.LicenseLimitFunctions, 1)))
	assert.NoError(t, h.Check(ctx, tx, licenseservice.Adding(base.LicenseLimitUsers, 1)))
	assert.ErrorIs(t, h.Check(ctx, tx, licenseservice.Adding(base.LicenseLimitPreviews, 1)),
		hperrors.ErrPlanLimitReached)
	assert.ErrorIs(t, h.Check(ctx, tx, licenseservice.Adding(base.LicenseLimitUsers, 2)),
		hperrors.ErrPlanLimitReached)
}

// A creation that adds several kinds takes their locks in base.AllLicenseLimits
// order, whatever order it names them in, so two of them cannot deadlock.
func TestLocksAreTakenInOneOrder(t *testing.T) {
	h := newHarness(t)

	err := h.Check(context.Background(), tx, &licenseservice.CheckReq{Adding: map[base.LicenseLimit]int{
		base.LicenseLimitPreviews:  1,
		base.LicenseLimitApps:      1,
		base.LicenseLimitFunctions: 1,
		base.LicenseLimitUsers:     1,
	}})

	assert.NoError(t, err)
	assert.Equal(t, []string{"lock users", "count users", "lock apps", "count apps",
		"lock functions", "count functions", "lock previews", "count previews"}, h.usage.asked)
}

// The first refusal ends the check: nothing more is locked.
func TestARefusalLocksNothingMore(t *testing.T) {
	h := newHarness(t)
	h.usage.counts[base.LicenseLimitUsers] = 2

	err := h.Check(context.Background(), tx, &licenseservice.CheckReq{Adding: map[base.LicenseLimit]int{
		base.LicenseLimitApps:  1,
		base.LicenseLimitUsers: 1,
	}})

	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	assert.Equal(t, []string{"lock users", "count users"}, h.usage.asked)
}

// The lock is the transaction's, so counting outside one would serialize
// nothing. That is a caller's mistake, refused on every plan so it shows on a
// paid one as well.
func TestACheckOutsideATransactionIsRefused(t *testing.T) {
	for name, store := range map[string]bool{"on Bee": false, "on a paid plan": true} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			if store {
				h.storeKey(t, payload())
			}

			err := h.Check(context.Background(), (*database.DB)(nil), licenseservice.Adding(base.LicenseLimitApps, 1))

			assert.ErrorIs(t, err, hperrors.ErrInternal)
			assert.Empty(t, h.usage.asked)
		})
	}
}

// Nodes are the swarm's: counted from docker, with no lock and no transaction
// needed. Every node counts, whatever its state.
func TestNodesAreCountedFromTheSwarm(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	h.docker.nodes = 1
	assert.NoError(t, h.Check(ctx, (*database.DB)(nil), licenseservice.Adding(base.LicenseLimitNodes, 1)))
	h.docker.nodes = 2
	assert.ErrorIs(t, h.Check(ctx, (*database.DB)(nil), licenseservice.Adding(base.LicenseLimitNodes, 1)),
		hperrors.ErrPlanLimitReached)

	assert.Equal(t, 2, h.docker.nodeCalls)
	assert.Empty(t, h.usage.asked)
}

// Bee adds no paid sign-in provider, whatever exists; a paid plan adds any.
// Neither counts: the ones that exist keep working.
func TestPaidSignInProvidersNeedAPaidPlan(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.usage.oauth = 3

	err := h.Check(ctx, (*database.DB)(nil), licenseservice.Adding(base.LicenseLimitSignInProviders, 1))
	assert.ErrorIs(t, err, hperrors.ErrPlanLimitReached)
	msg, transErr := hperrors.Wrap(err).Message(translation.LangEn)
	assert.NoError(t, transErr)
	assert.Equal(t, "The Bee plan does not include Google, Microsoft or OpenID Connect sign-in providers. "+
		"Upgrade to add one.", msg)

	h.storeKey(t, payload())
	assert.NoError(t, h.Check(ctx, (*database.DB)(nil), licenseservice.Adding(base.LicenseLimitSignInProviders, 1)))
	assert.Empty(t, h.usage.asked)
}

// The refusal carries Limit, Max and Plan on the wire, for the dashboard's
// upgrade prompt, and says the same in words.
func TestTheRefusalExposesLimitMaxAndPlan(t *testing.T) {
	h := newHarness(t)
	h.usage.counts[base.LicenseLimitApps] = 20

	err := h.Check(context.Background(), tx, licenseservice.Adding(base.LicenseLimitApps, 1))

	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, http.StatusPreconditionFailed, info.Status)
	assert.Equal(t, "ERR_PLAN_LIMIT_REACHED", info.Code)
	assert.Equal(t, map[string]any{"Limit": "apps", "Max": 20, "Plan": "bee"}, info.Params)
	assert.Equal(t, "The Bee plan allows 20 apps. Remove one, or upgrade.", info.Detail)
}

// A key's own limits apply, as its plan.
func TestAKeysLimitsApply(t *testing.T) {
	h := newHarness(t)
	h.storeKey(t, payload(func(p *license.Payload) { p.Plan = base.LicensePlanComb; p.Limits["users"] = 5 }))
	h.usage.counts[base.LicenseLimitUsers] = 5

	err := h.Check(context.Background(), tx, licenseservice.Adding(base.LicenseLimitUsers, 1))

	info, _ := hperrors.ParseError(err, translation.LangEn)
	assert.Equal(t, map[string]any{"Limit": "users", "Max": 5, "Plan": "comb"}, info.Params)
	assert.Equal(t, "The Comb plan allows 5 users. Remove one, or upgrade.", info.Detail)
}

// Adding nothing reads nothing.
func TestAddingNothingIsNotChecked(t *testing.T) {
	h := newHarness(t)

	assert.NoError(t, h.Check(context.Background(), (*database.DB)(nil), &licenseservice.CheckReq{}))
	assert.NoError(t, h.Check(context.Background(), (*database.DB)(nil),
		licenseservice.Adding(base.LicenseLimitApps, 0)))

	assert.Zero(t, h.settings.reads)
	assert.Empty(t, h.usage.asked)
}
