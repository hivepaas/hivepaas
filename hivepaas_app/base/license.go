package base

// LicensePlan is a plan: what the limits of an installation are. Bee is the free
// plan, the one an installation runs without a valid license key, and no key
// ever names it. The others are what a key can lift the limits to.
type LicensePlan string

const (
	LicensePlanBee    LicensePlan = "bee"
	LicensePlanComb   LicensePlan = "comb"
	LicensePlanHive   LicensePlan = "hive"
	LicensePlanColony LicensePlan = "colony"
	LicensePlanApiary LicensePlan = "apiary"
)

var (
	// AllLicensePlans run from the free plan up.
	AllLicensePlans = []LicensePlan{LicensePlanBee, LicensePlanComb, LicensePlanHive, LicensePlanColony,
		LicensePlanApiary}
)

// DisplayName is the plan's name as a person reads it, in messages and on the
// dashboard. A value no constant names, such as one read from a newer key, is
// shown as it is.
func (p LicensePlan) DisplayName() string {
	switch p {
	case LicensePlanBee:
		return "Bee"
	case LicensePlanComb:
		return "Comb"
	case LicensePlanHive:
		return "Hive"
	case LicensePlanColony:
		return "Colony"
	case LicensePlanApiary:
		return "Apiary"
	default:
		return string(p)
	}
}

// LicenseStatus is what the stored license key amounts to on this installation,
// now. Only valid, trial and grace, and moved until its grace runs out, put the
// key's limits in force; every other status leaves the installation on Bee.
type LicenseStatus string

const (
	// LicenseStatusNone is no key stored.
	LicenseStatusNone LicenseStatus = "none"
	// LicenseStatusValid is a key within its paid period.
	LicenseStatusValid LicenseStatus = "valid"
	// LicenseStatusTrial is a trial key within its period.
	LicenseStatusTrial LicenseStatus = "trial"
	// LicenseStatusGrace is a key past its period and within its grace days.
	LicenseStatusGrace LicenseStatus = "grace"
	// LicenseStatusExpired is a key past its grace days.
	LicenseStatusExpired LicenseStatus = "expired"
	// LicenseStatusInvalid is a key whose signatures or content do not verify.
	LicenseStatusInvalid LicenseStatus = "invalid"
	// LicenseStatusOtherInstallation is a valid key issued for other installations.
	LicenseStatusOtherInstallation LicenseStatus = "other-installation"
	// LicenseStatusMoved is a key the portal has since moved to another
	// installation. It holds until its grace runs out, counted from the move.
	LicenseStatusMoved LicenseStatus = "moved"
)

var (
	AllLicenseStatuses = []LicenseStatus{LicenseStatusNone, LicenseStatusValid, LicenseStatusTrial,
		LicenseStatusGrace, LicenseStatusExpired, LicenseStatusInvalid, LicenseStatusOtherInstallation,
		LicenseStatusMoved}
)

// LicenseLimit is a kind of thing a plan limits, and the name a license key's
// limits give it.
type LicenseLimit string

const (
	// LicenseLimitNodes counts the nodes of the swarm, whatever their state.
	LicenseLimitNodes LicenseLimit = "nodes"
	// LicenseLimitUsers counts the users active or pending.
	LicenseLimitUsers LicenseLimit = "users"
	// LicenseLimitApps counts the apps that are not functions, previews or apps
	// belonging to another.
	LicenseLimitApps LicenseLimit = "apps"
	// LicenseLimitFunctions counts the functions, with the same exclusions as apps.
	LicenseLimitFunctions LicenseLimit = "functions"
	// LicenseLimitPreviews counts the previews that exist, without the databases
	// they clone.
	LicenseLimitPreviews LicenseLimit = "previews"
	// LicenseLimitSignInProviders is the paid sign-in kinds (Google, Microsoft,
	// OpenID Connect). It is not a count: Bee allows none to be added, and every
	// other plan allows all.
	LicenseLimitSignInProviders LicenseLimit = "sign-in-providers"
)

var (
	// AllLicenseLimits is also the order a creation takes the plan-limit locks
	// in, so two creations that each touch two kinds cannot deadlock.
	AllLicenseLimits = []LicenseLimit{LicenseLimitNodes, LicenseLimitUsers, LicenseLimitApps,
		LicenseLimitFunctions, LicenseLimitPreviews, LicenseLimitSignInProviders}
)

// LicenseFeature is something a plan includes or leaves out, rather than counts:
// what a template requires and a key grants. A key names features, not plans
// alone, so which plans grant one can change without a template changing.
type LicenseFeature string

const (
	// LicenseFeatureManagedDatabases is creating apps from managed database
	// templates. What exists keeps running without it, and restoring is never
	// refused for want of it.
	LicenseFeatureManagedDatabases LicenseFeature = "managed-databases"
)

var (
	AllLicenseFeatures = []LicenseFeature{LicenseFeatureManagedDatabases}

	// BetaLicenseFeatures are what a beta installation has whatever its license:
	// free while HivePaaS is in beta. A stable installation has them from its key
	// alone.
	BetaLicenseFeatures = []LicenseFeature{LicenseFeatureManagedDatabases}
)

// LicenseRefreshResult is how a check with the portal for a newer key went.
type LicenseRefreshResult string

const (
	// LicenseRefreshNewer is a newer key for this installation, installed: a
	// renewal, a plan change, or a paid key replacing a trial.
	LicenseRefreshNewer LicenseRefreshResult = "newer"
	// LicenseRefreshMoved is the key moved to another installation since.
	LicenseRefreshMoved LicenseRefreshResult = "moved"
	// LicenseRefreshCurrent is no newer key: the one stored is the one in force.
	LicenseRefreshCurrent LicenseRefreshResult = "current"
	// LicenseRefreshUnknown is a key the portal did not issue, such as one signed
	// by hand or by the development keys.
	LicenseRefreshUnknown LicenseRefreshResult = "unknown"
	// LicenseRefreshNone is no key stored, so nothing to ask with: a first key is
	// pasted.
	LicenseRefreshNone LicenseRefreshResult = "none"
	// LicenseRefreshDisabled is the daily check switched off in the config.
	LicenseRefreshDisabled LicenseRefreshResult = "disabled"
	// LicenseRefreshFailed is a portal that could not be reached or refused to
	// answer, or a newer key that could not be installed.
	LicenseRefreshFailed LicenseRefreshResult = "failed"
)

var (
	AllLicenseRefreshResults = []LicenseRefreshResult{LicenseRefreshNewer, LicenseRefreshMoved,
		LicenseRefreshCurrent, LicenseRefreshUnknown, LicenseRefreshNone, LicenseRefreshDisabled,
		LicenseRefreshFailed}
)
