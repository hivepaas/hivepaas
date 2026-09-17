package hperrors

// Errors for app templates
var (
	ErrAppTemplateInvalid      = NewErr(ErrValueInvalid, "ERR_APP_TEMPLATE_INVALID")
	ErrAppTemplateParamInvalid = NewErr(ErrArgumentInvalid, "ERR_APP_TEMPLATE_PARAM_INVALID")
)
