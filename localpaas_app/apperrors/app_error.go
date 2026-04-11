package apperrors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	goerrors "github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/localpaas/localpaas/localpaas_app/infra/logging"
	"github.com/localpaas/localpaas/localpaas_app/pkg/translation"
)

type DisplayLevel string

const (
	DisplayLevelHigh   DisplayLevel = "high"
	DisplayLevelMedium DisplayLevel = "medium"
	DisplayLevelLow    DisplayLevel = "low"
)

const (
	errDetailsConsiderLong = 200
)

// AppError represents an error type to be used for any issue within the app.
// This error type is designed to be able to carry much extra information
// and ability to translate the error message into a specific language.
type AppError interface {
	error

	// WithCause sets cause of the error
	WithCause(cause error) AppError
	// WithParam sets a custom param (the param will be translated when build message)
	WithParam(k string, v any) AppError
	// WithParams sets custom params
	WithParams(map[string]any) AppError
	// WithNTParam sets a custom but non-translation param
	WithNTParam(k string, v any) AppError
	// WithExtraDetail sets extra detail
	WithExtraDetail(string, ...any) AppError
	// WithMsgLog sets log message (used for debug purpose)
	WithMsgLog(string, ...any) AppError

	// DisplayLevel get/set display level
	DisplayLevel() DisplayLevel
	WithDisplayLevel(DisplayLevel) AppError
	WithDisplayLevelHigh() AppError
	WithDisplayLevelMedium() AppError

	// FallbackToErrorMsg get/set fallback mode when translation missing
	FallbackToErrorMsg() bool
	WithFallbackToErrorMsg(flag bool) AppError

	// Message builds representation message
	Message(lang translation.Lang) (msg string, transErr error)
	// Build builds error info for JSON API recommendation
	Build(lang translation.Lang) *ErrorInfo

	// UnwrapTilRoot unwraps til the root error
	UnwrapTilRoot() error
}

// appError implements AppError interface
type appError struct {
	err                error
	cause              error
	params             map[string]any
	ntParams           map[string]any // non-translation params
	extraDetail        string
	msgLog             string
	displayLevel       DisplayLevel
	fallbackToErrorMsg bool // when translation missing
}

// Error implements `error` interface
func (e *appError) Error() string {
	return e.err.Error()
}

func (e *appError) WithCause(cause error) AppError {
	e.cause = cause
	return e
}

func (e *appError) WithParam(k string, v any) AppError {
	e.params[k] = v
	return e
}

func (e *appError) WithParams(m map[string]any) AppError {
	maps.Copy(e.params, m)
	return e
}

func (e *appError) WithNTParam(k string, v any) AppError {
	e.ntParams[k] = v
	return e
}

func (e *appError) WithExtraDetail(format string, args ...any) AppError {
	in := fmt.Sprintf(format, args...)
	if e.extraDetail == "" {
		e.extraDetail = in
		return e
	}
	e.extraDetail = fmt.Sprintf("%s\n%s", e.extraDetail, in)
	return e
}

func (e *appError) WithMsgLog(format string, args ...any) AppError {
	in := fmt.Sprintf(format, args...)
	if e.msgLog == "" {
		e.msgLog = in
		return e
	}
	e.msgLog = fmt.Sprintf("%s\n%s", e.msgLog, in)
	return e
}

func (e *appError) DisplayLevel() DisplayLevel {
	return e.displayLevel
}

func (e *appError) WithDisplayLevel(level DisplayLevel) AppError {
	e.displayLevel = level
	return e
}

func (e *appError) WithDisplayLevelHigh() AppError {
	e.displayLevel = DisplayLevelHigh
	return e
}

func (e *appError) WithDisplayLevelMedium() AppError {
	e.displayLevel = DisplayLevelMedium
	return e
}

func (e *appError) FallbackToErrorMsg() bool {
	return e.fallbackToErrorMsg
}

func (e *appError) WithFallbackToErrorMsg(flag bool) AppError {
	e.fallbackToErrorMsg = flag
	return e
}

// Build - builder (status, code, title, detail)
func (e *appError) Build(lang translation.Lang) *ErrorInfo {
	errInfo := &ErrorInfo{}

	errInfo.Status = e.getMappingStatus()
	if errInfo.Status == 0 {
		errInfo.Status = http.StatusInternalServerError
		errInfo.Code = ErrInternalServer.Error()
	} else {
		errInfo.Code = e.err.Error()
	}

	detail, transErr := e.Message(lang)
	if transErr != nil {
		// This is not error, just notify dev team about missing translation
		notifyTranslationMissing(transErr, lang)
		if detail == "" && e.fallbackToErrorMsg {
			detail = e.err.Error()
		}
	}
	if e.extraDetail != "" {
		detail = detail + "\n\n" + e.extraDetail
	}

	errInfo.Title = http.StatusText(errInfo.Status)
	errInfo.Detail = detail
	errInfo.DebugLog = e.msgLog
	errInfo.DisplayLevel = e.displayLevel
	if len(detail) > errDetailsConsiderLong || errors.Is(e.err, ErrInfra) { // TODO: rune count?
		errInfo.DisplayLevel = DisplayLevelHigh
	}
	if e.cause != nil {
		errInfo.Cause = e.cause.Error()
	} else {
		errInfo.Cause = e.err.Error()
	}
	errInfo.StackTrace = e.StackTrace()

	return errInfo
}

func (e *appError) Message(lang translation.Lang) (msg string, transErr error) {
	params := make(map[string]any, len(e.params)+len(e.ntParams))
	maps.Copy(params, e.ntParams)
	for k, v := range e.params {
		vAsStr, ok := v.(string)
		if !ok {
			params[k] = v
			continue
		}
		if translated, err := translation.Localize(lang, vAsStr); err != nil {
			transErr = multierror.Append(transErr, err)
			params[k] = v
		} else {
			params[k] = translated
		}
	}

	msgID := e.UnwrapTilRoot().Error()
	missingTranslation := false
	if strings.HasPrefix(msgID, "ERR_") {
		var err error
		msg, err = translation.LocalizeEx(lang, msgID, params)
		if err != nil {
			transErr = multierror.Append(transErr, err)
			missingTranslation = true
		}
	} else {
		missingTranslation = true
	}

	if missingTranslation {
		if e.fallbackToErrorMsg {
			msg = e.Error()
		} else {
			msg, _ = translation.Localize(lang, ErrInternalServer.Error()) // Show error 500
		}
	}
	return msg, transErr //nolint:wrapcheck
}

// Is - implements errors.Is.
// This returns true if either the inner error or the cause satisfies.
func (e *appError) Is(err error) bool {
	if errors.Is(e.err, err) {
		return true
	}
	if e.cause != nil {
		return errors.Is(e.cause, err)
	}
	return false
}

// Unwrap - implements errors.Unwrap
func (e *appError) Unwrap() error {
	return e.err
}

// UnwrapTilRoot keeps unwrapping until the root error
func (e *appError) UnwrapTilRoot() error {
	lastErr := e.err
	for {
		err := errorUnwrap(lastErr)
		if err == nil {
			return lastErr
		}
		lastErr = err
	}
}

func errorUnwrap(err error) error {
	u, ok := err.(interface{ Unwrap() error })
	if ok {
		return u.Unwrap() //nolint:wrapcheck
	}
	u2, ok := err.(interface{ Unwrap() []error })
	if ok {
		res := u2.Unwrap()
		if len(res) > 0 {
			return res[0]
		}
	}
	return nil
}

func (e *appError) StackTrace() string {
	var errWithStack *goerrors.Error
	if errors.As(e.err, &errWithStack) {
		return errWithStack.ErrorStack()
	}

	return ""
}

func (e *appError) getMappingStatus() int {
	// errorStatusMap[error] with an unhashable input error object
	// can cause panic. We recover from panic and return 0.
	defer func() {
		_ = recover()
	}()
	err := e.err
	for {
		if err == nil {
			return 0
		}
		if status, ok := errorStatusMap[err]; ok {
			return status
		}
		err = errors.Unwrap(err)
	}
}

func notifyTranslationMissing(e error, _ translation.Lang) {
	// the error format is something like this:
	// 1 error occurred:
	// \* message "ERR_BAD_REQUEST" not found in language "en"
	// It does have a line break and the actual error starts after '*',
	// so let's take the substring after '*' for the logging.
	_, errMsg, _ := strings.Cut(e.Error(), "* ")
	logging.Errorf("%s", errMsg)
}

func New(err error) AppError {
	if err == nil {
		return nil
	}
	var e *appError
	if errors.As(err, &e) {
		return e // already is a AppError, no need to wrap
	}
	return &appError{
		ntParams:           map[string]any{},
		params:             map[string]any{},
		fallbackToErrorMsg: true,
		err:                goerrors.Wrap(err, 1),
	}
}
