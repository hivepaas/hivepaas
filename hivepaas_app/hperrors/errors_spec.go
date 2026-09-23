package hperrors

// Errors for configuration spec export
var (
	ErrSpecMountTargetDuplicated   = NewErr(ErrValueInvalid, "ERR_SPEC_MOUNT_TARGET_DUPLICATED")
	ErrSpecSettingTypeUnclassified = NewErr(ErrInternal, "ERR_SPEC_SETTING_TYPE_UNCLASSIFIED")
	ErrSpecPassphraseRequired      = NewErr(ErrPreconditionRequired, "ERR_SPEC_PASSPHRASE_REQUIRED")
	ErrSpecSecretsModeInvalid      = NewErr(ErrArgumentInvalid, "ERR_SPEC_SECRETS_MODE_INVALID")
	ErrSpecBlockUnsupported        = NewErr(ErrUnsupported, "ERR_SPEC_BLOCK_UNSUPPORTED")
	ErrSpecBlockInvalid            = NewErr(ErrArgumentInvalid, "ERR_SPEC_BLOCK_INVALID")
)

// Errors for configuration spec import
var (
	ErrSpecPassphraseInvalid      = NewErr(ErrArgumentInvalid, "ERR_SPEC_PASSPHRASE_INVALID")
	ErrSpecBundleInvalid          = NewErr(ErrArgumentInvalid, "ERR_SPEC_BUNDLE_INVALID")
	ErrSpecAPIVersionUnsupported  = NewErr(ErrUnsupported, "ERR_SPEC_API_VERSION_UNSUPPORTED")
	ErrSpecImportScopeNotInBundle = NewErr(ErrArgumentInvalid, "ERR_SPEC_IMPORT_SCOPE_NOT_IN_BUNDLE")
)
