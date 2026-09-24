package hperrors

// System errors
var (
	ErrUpdateVerMismatched      = NewErr(ErrMismatch, "ERR_UPDATE_VER_MISMATCHED")
	ErrVersionNotNewer          = NewErr(ErrPreconditionFailed, "ERR_VERSION_NOT_NEWER")
	ErrReleaseSignatureInvalid  = NewErr(ErrPreconditionFailed, "ERR_RELEASE_SIGNATURE_INVALID")
	ErrReleaseSigningKeyInvalid = NewErr(ErrInternal, "ERR_RELEASE_SIGNING_KEY_INVALID")
	ErrReleaseInfoInvalid       = NewErr(ErrPreconditionFailed, "ERR_RELEASE_INFO_INVALID")
	// A component the release says must not cross a major version would have to.
	ErrSystemUpdateBlocked = NewErr(ErrPreconditionFailed, "ERR_SYSTEM_UPDATE_BLOCKED")
	// The update moves a component from the database backup, and was asked to skip it.
	ErrSystemUpdateNeedsBackup = NewErr(ErrPreconditionFailed, "ERR_SYSTEM_UPDATE_NEEDS_BACKUP")
)
