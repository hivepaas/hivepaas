package hperrors

// Errors for session
var (
	ErrNoSession                   = NewErr(ErrUnauthorized, "ERR_NO_SESSION")
	ErrSessionJWTInvalid           = NewErr(ErrUnauthorized, "ERR_SESSION_JWT_INVALID")
	ErrSessionJWTExpired           = NewErr(ErrUnauthorized, "ERR_SESSION_JWT_EXPIRED")
	ErrSessionAPIKeyInvalid        = NewErr(ErrUnauthorized, "ERR_SESSION_API_KEY_INVALID")
	ErrSessionRefreshTokenRequired = NewErr(ErrUnauthorized, "ERR_SESSION_REFRESH_TOKEN_REQUIRED")
	ErrSSORequired                 = NewErr(ErrUnauthorized, "ERR_SSO_REQUIRED")
	ErrLoginInputInvalid           = NewErr(ErrUnauthorized, "ERR_LOGIN_INPUT_INVALID")
	ErrPasswordMismatched          = NewErr(ErrUnauthorized, "ERR_PASSWORD_MISMATCHED")
	ErrPasscodeMismatched          = NewErr(ErrUnauthorized, "ERR_PASSCODE_MISMATCHED")
	ErrTooManyLoginFailures        = NewErr(ErrUnauthorized, "ERR_TOO_MANY_LOGIN_FAILURES")
	ErrTooManyPasscodeAttempts     = NewErr(ErrUnauthorized, "ERR_TOO_MANY_PASSCODE_ATTEMPTS")
	ErrOAuthUserEmailNotReturned   = NewErr(ErrPreconditionFailed, "ERR_OAUTH_USER_EMAIL_NOT_RETURNED")
	// ErrOAuthEmailNotVerified refuses a sign-in with a provider that does not vouch the person proved they hold the
	// email: matched to a user by it, it would let whoever typed it in sign in as them.
	ErrOAuthEmailNotVerified = NewErr(ErrUnauthorized, "ERR_OAUTH_EMAIL_NOT_VERIFIED")
	// ErrOAuthUserNotFound is a sign-in with a provider whose email no user has: SSO signs in users that exist.
	ErrOAuthUserNotFound = NewErr(ErrUnauthorized, "ERR_OAUTH_USER_NOT_FOUND")
)

// Errors for user
var (
	ErrUserNotFound             = NewErr(ErrNotFound, "ERR_USER_NOT_FOUND")
	ErrUserUnavailable          = NewErr(ErrUnauthorized, "ERR_USER_UNAVAILABLE")
	ErrUsernameUnavailable      = NewErr(ErrUnavailable, "ERR_USERNAME_UNAVAILABLE")
	ErrUserDemoUnauthorized     = NewErr(ErrUnauthorized, "ERR_USER_DEMO_UNAUTHORIZED")
	ErrUserStatusNotAllowAction = NewErr(ErrNotAllowed, "ERR_USER_STATUS_NOT_ALLOW_ACTION")
	ErrUserNotCompleteMFASetup  = NewErr(ErrPreconditionFailed, "ERR_USER_NOT_COMPLETE_MFA_SETUP")

	// ErrUserAdminSecurityOptionWeak refuses to leave an admin authenticating with
	// a password and nothing else. See base.SecurityOptionAllowedForRole.
	ErrUserAdminSecurityOptionWeak          = NewErr(ErrPreconditionFailed, "ERR_USER_ADMIN_SECURITY_OPTION_WEAK")
	ErrEmailUnavailable                     = NewErr(ErrUnavailable, "ERR_EMAIL_UNAVAILABLE")
	ErrEmailChangeUnallowed                 = NewErr(ErrNotAllowed, "ERR_EMAIL_CHANGE_UNALLOWED")
	ErrUserNotHavePermissionOnResource      = NewErr(ErrUnauthorized, "ERR_USER_NOT_HAVE_PERMISSION_ON_RESOURCE")
	ErrUserNotHavePermissionOnRevealSecrets = NewErr(ErrUnauthorized,
		"ERR_USER_NOT_HAVE_PERMISSION_ON_REVEAL_SECRETS")
	// ErrRevealSecretsDisabled is the operator's switch, not the caller's
	// permissions: it is off for everyone, admins included, until the host
	// configuration turns it on.
	ErrRevealSecretsDisabled               = NewErr(ErrNotAllowed, "ERR_REVEAL_SECRETS_DISABLED")
	ErrUserNotHavePermissionOnCreateAPIKey = NewErr(ErrUnauthorized,
		"ERR_USER_NOT_HAVE_PERMISSION_ON_CREATE_API_KEY")
)

// Errors for api client
var (
	// ErrAPIKeyInvalid is a key nobody knows, switched off, or with the wrong
	// secret: the caller is not authenticated, as with no credentials at all.
	ErrAPIKeyInvalid = NewErr(ErrUnauthorized, "ERR_API_KEY_INVALID")
)
