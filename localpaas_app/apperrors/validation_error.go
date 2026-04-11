package apperrors

import (
	"net/http"
	"strings"

	"github.com/hashicorp/go-multierror"
	validation "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/pkg/translation"
)

const (
	errVldStatusCode   = http.StatusBadRequest
	errVldDisplayLevel = DisplayLevelMedium

	errKeyVldCommonDetail = "ERR_VLD_COMMON_DETAIL"
)

// ValidationError generic validation error, typically depending on the validation library
// we use, this error type is a wrapper of the errors returned by that lib.
type ValidationError struct {
	err validation.Error
}

// ValidationErrors bunch of ValidationError
type ValidationErrors []*ValidationError

// Build - build list of ErrorInfo for this error
func (e ValidationErrors) Build(lang translation.Lang) *ErrorInfo {
	vldErrs := make([]*InnerErrorInfo, 0, len(e))
	var transErr error
	for i := range e {
		errInfo, err := e[i].build(lang)
		if err != nil {
			transErr = multierror.Append(transErr, err)
		}
		vldErrs = append(vldErrs, errInfo)
	}
	detail, err := translation.Localize(lang, errKeyVldCommonDetail)
	if err != nil {
		transErr = multierror.Append(transErr, err)
	}
	if transErr != nil {
		// This is not error, just notify dev team about missing translation
		notifyTranslationMissing(transErr, lang)
	}

	return &ErrorInfo{
		Title:        http.StatusText(errVldStatusCode),
		Status:       errVldStatusCode,
		Code:         ErrValidation.Error(),
		Detail:       detail,
		DisplayLevel: errVldDisplayLevel,
		InnerErrors:  vldErrs,
	}
}

func (e ValidationErrors) Error() string {
	var sb strings.Builder
	for i := range e {
		if i > 0 {
			sb.WriteString("\n")
		}
		detail, _, _ := e[i].message(translation.LangEn)
		sb.WriteString(detail)
	}
	return sb.String()
}

// Error - error.Error
func (e *ValidationError) Error() string {
	return e.err.Error()
}

// Build - build ErrorInfo for this error
func (e *ValidationError) build(lang translation.Lang) (*InnerErrorInfo, error) {
	msg, field, transErr := e.message(lang)
	errInfo := &InnerErrorInfo{
		Code:    ErrValidation.Error(),
		Path:    field,
		Message: msg,
		Cause:   e.err.Error(),
	}

	return errInfo, transErr
}

func (e *ValidationError) message(lang translation.Lang) (msg string, source string, transErr error) {
	f := e.err.Field()
	if f != nil {
		source = f.PathString(true, "/")
	}

	inErr := e.err
	params := inErr.ParamsWithFormatter()
	var err error

	if customKey := inErr.CustomKey(); customKey != nil {
		customKeyStr, _ := customKey.(string)
		msg, err = translation.LocalizeEx(lang, customKeyStr, params)
		if err != nil {
			transErr = multierror.Append(transErr, err)
		}
	}
	if msg == "" {
		msg, err = inErr.BuildDetail()
		if err != nil {
			transErr = multierror.Append(transErr, err)
		}
	}

	return msg, source, transErr //nolint:wrapcheck
}

// NewValidationErrors wraps errors from 3rd party lib
func NewValidationErrors(re validation.Errors) ValidationErrors {
	if len(re) == 0 {
		return nil
	}

	errs := make([]*ValidationError, len(re))
	for i := range re {
		errs[i] = &ValidationError{err: re[i]}
	}
	return errs
}
