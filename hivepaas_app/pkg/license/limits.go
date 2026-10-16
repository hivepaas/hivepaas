package license

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Limits is the most of each kind a plan allows. A kind with no entry has no
// limit.
type Limits map[base.LicenseLimit]int

// Max is the most of limit allowed, and whether there is a most at all.
func (l Limits) Max(limit base.LicenseLimit) (int, bool) {
	maxCount, limited := l[limit]
	return maxCount, limited
}

// The free plan's limits, as the licensing design's plan table sets them.
const (
	beeNodes     = 2
	beeUsers     = 2
	beeApps      = 20
	beeFunctions = 20
	beePreviews  = 2
)

// BeeLimits are the free plan's. They are compiled in, never in a key, and
// Evaluate gives each State a copy of its own. Bee allows no paid sign-in
// provider to be added.
var BeeLimits = Limits{
	base.LicenseLimitNodes:           beeNodes,
	base.LicenseLimitUsers:           beeUsers,
	base.LicenseLimitApps:            beeApps,
	base.LicenseLimitFunctions:       beeFunctions,
	base.LicenseLimitPreviews:        beePreviews,
	base.LicenseLimitSignInProviders: 0,
}

// LimitsOf is the limits a key puts in force: those of its limits this binary
// knows.
func LimitsOf(p *Payload) Limits {
	limits := Limits{}
	for _, limit := range base.AllLicenseLimits {
		if !setByKey(limit) {
			continue
		}
		if maxCount, found := p.Limits[string(limit)]; found {
			limits[limit] = maxCount
		}
	}
	return limits
}

// setByKey reports whether a key's limits can set limit.
func setByKey(limit base.LicenseLimit) bool {
	switch limit {
	case base.LicenseLimitNodes, base.LicenseLimitUsers, base.LicenseLimitApps, base.LicenseLimitFunctions,
		base.LicenseLimitPreviews:
		return true
	case base.LicenseLimitSignInProviders:
		// Not a count: no paid sign-in provider may be added on Bee, and every
		// paid plan has them all. A number in a key could only mean something
		// it cannot enforce.
		return false
	}
	return false
}

// IsPaidSignIn reports whether a sign-in provider of kind is one the free plan
// leaves out: signing in with a company's identity provider rather than with a
// code host.
func IsPaidSignIn(kind base.OAuthKind) bool {
	switch kind {
	case base.OAuthKindGoogle, base.OAuthKindMicrosoftOnline, base.OAuthKindOpenIDConnect:
		return true
	case base.OAuthKindGithub, base.OAuthKindGithubApp, base.OAuthKindGitlab, base.OAuthKindGitea:
		fallthrough
	default:
		return false
	}
}

// LimitNameMsgID is the message id of what ERR_PLAN_LIMIT_REACHED calls limit:
// the translatable LimitName it is raised with.
func LimitNameMsgID(limit base.LicenseLimit) string {
	switch limit {
	case base.LicenseLimitNodes:
		return "ERR_PLAN_LIMIT_NODES"
	case base.LicenseLimitUsers:
		return "ERR_PLAN_LIMIT_USERS"
	case base.LicenseLimitApps:
		return "ERR_PLAN_LIMIT_APPS"
	case base.LicenseLimitFunctions:
		return "ERR_PLAN_LIMIT_FUNCTIONS"
	case base.LicenseLimitPreviews:
		return "ERR_PLAN_LIMIT_PREVIEWS"
	case base.LicenseLimitSignInProviders:
		return "ERR_PLAN_LIMIT_SIGN_IN_PROVIDERS"
	}
	return ""
}
