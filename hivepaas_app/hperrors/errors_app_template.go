package hperrors

// Errors for app templates
var (
	ErrAppTemplateInvalid      = NewErr(ErrValueInvalid, "ERR_APP_TEMPLATE_INVALID")
	ErrAppTemplateParamInvalid = NewErr(ErrArgumentInvalid, "ERR_APP_TEMPLATE_PARAM_INVALID")

	ErrAppTemplateVersionNotFound    = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VERSION_NOT_FOUND")
	ErrAppTemplateVersionDeprecated  = NewErr(ErrNotAllowed, "ERR_APP_TEMPLATE_VERSION_DEPRECATED")
	ErrAppTemplateVariantUnavailable = NewErr(ErrNotFound, "ERR_APP_TEMPLATE_VARIANT_UNAVAILABLE")
)
