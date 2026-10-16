package license

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func TestLimitsMax(t *testing.T) {
	limits := Limits{base.LicenseLimitNodes: 5, base.LicenseLimitSignInProviders: 0}

	maxCount, limited := limits.Max(base.LicenseLimitNodes)
	assert.True(t, limited)
	assert.Equal(t, 5, maxCount)

	maxCount, limited = limits.Max(base.LicenseLimitSignInProviders)
	assert.True(t, limited, "a limit of 0 is a limit")
	assert.Equal(t, 0, maxCount)

	_, limited = limits.Max(base.LicenseLimitApps)
	assert.False(t, limited, "a limit left out is no limit")
}

// Bee limits every kind, so none is left out by mistake.
func TestBeeLimits(t *testing.T) {
	assert.Equal(t, Limits{
		base.LicenseLimitNodes:           2,
		base.LicenseLimitUsers:           2,
		base.LicenseLimitApps:            20,
		base.LicenseLimitFunctions:       20,
		base.LicenseLimitPreviews:        2,
		base.LicenseLimitSignInProviders: 0,
	}, BeeLimits)
	for _, limit := range base.AllLicenseLimits {
		_, limited := BeeLimits.Max(limit)
		assert.True(t, limited, limit)
	}
}

func TestLimitsOf(t *testing.T) {
	limits := LimitsOf(&Payload{Limits: map[string]int{
		"nodes": 25, "users": 30, "previews": 0,
		// Never limited by a key.
		"sign-in-providers": 0,
		// From a newer portal.
		"gpus": 4,
	}})
	assert.Equal(t, Limits{base.LicenseLimitNodes: 25, base.LicenseLimitUsers: 30, base.LicenseLimitPreviews: 0}, limits)

	assert.Empty(t, LimitsOf(&Payload{}), "a key with no limits limits nothing")
}

func TestIsPaidSignIn(t *testing.T) {
	paid := map[base.OAuthKind]bool{
		base.OAuthKindGithub:          false,
		base.OAuthKindGithubApp:       false,
		base.OAuthKindGitlab:          false,
		base.OAuthKindGitea:           false,
		base.OAuthKindGoogle:          true,
		base.OAuthKindMicrosoftOnline: true,
		base.OAuthKindOpenIDConnect:   true,
	}
	for _, kind := range base.AllOAuthKinds {
		want, decided := paid[kind]
		if !assert.True(t, decided, "kind %q: decide whether Bee allows it, here and in IsPaidSignIn", kind) {
			continue
		}
		assert.Equal(t, want, IsPaidSignIn(kind), kind)
	}
	assert.False(t, IsPaidSignIn("saml"), "a kind no constant names is not refused")
}

// Every limit has a name for ERR_PLAN_LIMIT_REACHED, and the message reads.
func TestPlanLimitMessages(t *testing.T) {
	message := func(limit base.LicenseLimit, plan base.LicensePlan, maxCount int) string {
		msg, err := hperrors.Wrap(hperrors.ErrPlanLimitReached).
			WithParam("Limit", string(limit)).
			WithParam("Max", maxCount).
			WithParam("Plan", string(plan)).
			WithTransParam("LimitName", LimitNameMsgID(limit)).
			WithParam("PlanName", plan.DisplayName()).
			Message(translation.LangEn)
		assert.NoError(t, err, limit)
		return msg
	}

	for _, limit := range base.AllLicenseLimits {
		assert.NotEmpty(t, LimitNameMsgID(limit), limit)
		maxCount, _ := BeeLimits.Max(limit)
		assert.NotContains(t, message(limit, base.LicensePlanBee, maxCount), "ERR_", limit)
	}

	assert.Equal(t, "The Bee plan allows 20 apps. Remove one, or upgrade.",
		message(base.LicenseLimitApps, base.LicensePlanBee, 20))
	assert.Equal(t, "The Bee plan allows 2 previews at a time. Remove one, or upgrade.",
		message(base.LicenseLimitPreviews, base.LicensePlanBee, 2))
	assert.Equal(t, "The Comb plan allows 5 nodes. Remove one, or upgrade.",
		message(base.LicenseLimitNodes, base.LicensePlanComb, 5))
	// Removing one would free nothing: Bee has none of these to give.
	assert.Equal(t, "The Bee plan does not include Google, Microsoft or OpenID Connect sign-in providers. "+
		"Upgrade to add one.", message(base.LicenseLimitSignInProviders, base.LicensePlanBee, 0))
}
