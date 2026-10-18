package hperrors

// Errors for plans and license keys.
//
// They are 412, never 403: the dashboard reads 403 as a permission the person
// lacks, and none of these is about who is asking.
var (
	// ErrPlanLimitReached refuses a creation that would take a limited kind past
	// the limit of the plan in force. The dashboard turns it into an upgrade prompt.
	ErrPlanLimitReached = NewErr(ErrPreconditionFailed, "ERR_PLAN_LIMIT_REACHED")
	// ErrPlanFeatureRequired refuses creating what needs a feature the plan in
	// force does not include, such as an app from a managed database template.
	ErrPlanFeatureRequired = NewErr(ErrPreconditionFailed, "ERR_PLAN_FEATURE_REQUIRED")
	// ErrLicenseInvalid is a license key whose signatures or content do not
	// verify. Its detail says what failed, never what the key holds: a key carries
	// the customer's email.
	ErrLicenseInvalid = NewErr(ErrValueInvalid, "ERR_LICENSE_INVALID")
	// ErrLicenseOtherInstallation refuses a valid key issued for another
	// installation, and names this installation's ID so it can be asked for.
	ErrLicenseOtherInstallation = NewErr(ErrPreconditionFailed, "ERR_LICENSE_OTHER_INSTALLATION")
	// ErrLicenseExpired refuses a valid key whose grace has run out.
	ErrLicenseExpired = NewErr(ErrPreconditionFailed, "ERR_LICENSE_EXPIRED")
	// ErrLicenseSigningKeyInvalid is a compiled-in license public key that does
	// not parse: a build mistake, not something a person can fix.
	ErrLicenseSigningKeyInvalid = NewErr(ErrInternal, "ERR_LICENSE_SIGNING_KEY_INVALID")
	// ErrLicensePortalUnavailable is a license portal that could not be reached,
	// or that is down or busy: asking again later may work.
	ErrLicensePortalUnavailable = NewErr(ErrServiceUnavailable, "ERR_LICENSE_PORTAL_UNAVAILABLE")
	// ErrLicensePortalRefused is a portal that answered a refresh with a refusal,
	// such as for a key that does not name this installation. Its detail is the
	// portal's code; asking again will not change the answer.
	ErrLicensePortalRefused = NewErr(ErrPreconditionFailed, "ERR_LICENSE_PORTAL_REFUSED")
)
