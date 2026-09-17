package hperrors

// Errors for app templates
var (
	ErrAppTemplateInvalid      = NewErr(ErrValueInvalid, "ERR_APP_TEMPLATE_INVALID")
	ErrAppTemplateParamInvalid = NewErr(ErrArgumentInvalid, "ERR_APP_TEMPLATE_PARAM_INVALID")

	ErrAppTemplateVersionNotFound    = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VERSION_NOT_FOUND")
	ErrAppTemplateVersionDeprecated  = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_VERSION_DEPRECATED")
	ErrAppTemplateVariantUnavailable = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VARIANT_UNAVAILABLE")

	ErrAppTemplateFileTooLarge = NewErr(ErrTooBig, "ERR_APP_TEMPLATE_FILE_TOO_LARGE")

	ErrAppTemplatesUnavailable       = NewErr(ErrUnavailable, "ERR_APP_TEMPLATES_UNAVAILABLE")
	ErrAppTemplateNotFound           = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_NOT_FOUND")
	ErrAppTemplateVerificationFailed = NewErr(ErrPreconditionFailed, "ERR_APP_TEMPLATE_VERIFICATION_FAILED")
	ErrAppTemplateIncompatible       = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_INCOMPATIBLE")
)
