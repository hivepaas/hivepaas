package hperrors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	goerrors "github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
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

// HPError represents an error type to be used for any issue within the app.
// This error type is designed to be able to carry much extra information
// and ability to translate the error message into a specific language.
type HPError interface {
	error

	// WithCause sets cause of the error
	WithCause(cause error) HPError
	// WithParam sets a custom param (the param will be translated when build message)
	WithParam(k string, v any) HPError
	// WithParams sets custom params
	WithParams(map[string]any) HPError
	// WithNTParam sets a custom but non-translation param
	WithNTParam(k string, v any) HPError
	// WithExtraDetail sets extra detail
	WithExtraDetail(string, ...any) HPError
	// WithMsgLog sets log message (used for debug purpose)
	WithMsgLog(string, ...any) HPError

	// DisplayLevel get/set display level
	DisplayLevel() DisplayLevel
	WithDisplayLevel(DisplayLevel) HPError
	WithDisplayLevelHigh() HPError
	WithDisplayLevelMedium() HPError

	// FallbackToErrorMsg get/set fallback mode when translation missing
	FallbackToErrorMsg() bool
	WithFallbackToErrorMsg(flag bool) HPError

	// StatusCode gets status code of error
	StatusCode() int
	// Message builds representation message
	Message(lang translation.Lang) (msg string, transErr error)
	// Build builds error info for JSON API recommendation
	Build(lang translation.Lang) *ErrorInfo
}

// hpError implements HPError interface
type hpError struct {
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
func (e *hpError) Error() string {
	return e.err.Error()
}

func (e *hpError) WithCause(cause error) HPError {
	e.cause = cause
	return e
}

func (e *hpError) WithParam(k string, v any) HPError {
	e.params[k] = v
	return e
}

func (e *hpError) WithParams(m map[string]any) HPError {
	maps.Copy(e.params, m)
	return e
}

func (e *hpError) WithNTParam(k string, v any) HPError {
	e.ntParams[k] = v
	return e
}

func (e *hpError) WithExtraDetail(format string, args ...any) HPError {
	in := fmt.Sprintf(format, args...)
	if e.extraDetail == "" {
		e.extraDetail = in
		return e
	}
	e.extraDetail = fmt.Sprintf("%s\n%s", e.extraDetail, in)
	return e
}

func (e *hpError) WithMsgLog(format string, args ...any) HPError {
	in := fmt.Sprintf(format, args...)
	if e.msgLog == "" {
		e.msgLog = in
		return e
	}
	e.msgLog = fmt.Sprintf("%s\n%s", e.msgLog, in)
	return e
}

func (e *hpError) DisplayLevel() DisplayLevel {
	return e.displayLevel
}

func (e *hpError) WithDisplayLevel(level DisplayLevel) HPError {
	e.displayLevel = level
	return e
}

func (e *hpError) WithDisplayLevelHigh() HPError {
	e.displayLevel = DisplayLevelHigh
	return e
}

func (e *hpError) WithDisplayLevelMedium() HPError {
	e.displayLevel = DisplayLevelMedium
	return e
}

func (e *hpError) FallbackToErrorMsg() bool {
	return e.fallbackToErrorMsg
}

func (e *hpError) WithFallbackToErrorMsg(flag bool) HPError {
	e.fallbackToErrorMsg = flag
	return e
}

// Build - builder (status, code, title, detail)
func (e *hpError) Build(lang translation.Lang) *ErrorInfo {
	errInfo := &ErrorInfo{}

	errInfo.Status = e.getMappingStatus()
	if errInfo.Status == 0 {
		errInfo.Status = http.StatusInternalServerError
		errInfo.Code = ErrInternal.Error()
	} else {
		errInfo.Code = getErrorCode(e.err)
	}

	detail, transErr := e.getMessage(errInfo.Code, lang)
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

func (e *hpError) StatusCode() int {
	return e.getMappingStatus()
}

func (e *hpError) Message(lang translation.Lang) (msg string, transErr error) {
	return e.getMessage("", lang)
}

func (e *hpError) getMessage(msgID string, lang translation.Lang) (msg string, transErr error) {
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

	if msgID == "" {
		msgID = getMessageID(e.err)
	}

	missingTranslation := false
	if msgID != "" {
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
			msg, _ = translation.Localize(lang, ErrInternal.Error()) // Show error 500
		}
	}
	return msg, transErr //nolint:wrapcheck
}

// Is - implements errors.Is.
// This returns true if either the inner error or the cause satisfies.
func (e *hpError) Is(err error) bool {
	if errors.Is(e.err, err) {
		return true
	}
	if e.cause != nil {
		return errors.Is(e.cause, err)
	}
	return false
}

// Unwrap - implements errors.Unwrap
func (e *hpError) Unwrap() error {
	return e.err
}

func (e *hpError) StackTrace() string {
	return GetErrorStackTrace(e.err)
}

func (e *hpError) getMappingStatus() int {
	baseErr := getBaseError(e.err)
	if baseErr != nil {
		return errorStatusMap[baseErr]
	}
	return http.StatusInternalServerError
}

func getBaseError(err error) error {
	// errorStatusMap[error] with an unhashable input error object
	// can cause panic. We recover from panic and return 0.
	defer func() {
		_ = recover()
	}()
	if err == nil {
		return nil
	}
	if _, ok := errorStatusMap[err]; ok {
		return err
	}
	u, ok := err.(interface{ Unwrap() error })
	if ok {
		return getBaseError(u.Unwrap())
	}
	u2, ok := err.(interface{ Unwrap() []error })
	if ok {
		for _, err := range u2.Unwrap() {
			if baseErr := getBaseError(err); baseErr != nil {
				return baseErr
			}
		}
	}
	return nil
}

func getErrorCode(err error) string {
	return getMessageID(err)
}

func getMessageID(err error) (msg string) {
	if err == nil {
		return ""
	}
	if msgID := err.Error(); isValidMessageID(msgID) {
		return msgID
	}
	u, ok := err.(interface{ Unwrap() error })
	if ok {
		return getMessageID(u.Unwrap())
	}
	u2, ok := err.(interface{ Unwrap() []error })
	if ok {
		errs := u2.Unwrap()
		for i := len(errs) - 1; i >= 0; i-- {
			if msgID := getMessageID(errs[i]); msgID != "" {
				return msgID
			}
		}
	}
	return ""
}

// isValidMessageID check if a string is a message ID (ERR_UPPERCASE_WORDS)
func isValidMessageID(s string) bool {
	if !strings.HasPrefix(s, "ERR_") {
		return false
	}
	for _, ch := range s {
		if ch != '_' && !('0' <= ch && ch <= '9') && !('A' <= ch && ch <= 'Z') { //nolint:staticcheck
			return false
		}
	}
	return true
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

func Wrap(err error) HPError {
	if err == nil {
		return nil
	}

	e, ok := errors.AsType[*hpError](err)
	if !ok {
		return &hpError{
			ntParams:           map[string]any{},
			params:             map[string]any{},
			fallbackToErrorMsg: true,
			err:                goerrors.Wrap(err, 1),
		}
	}

	// Deliberately an identity check rather than errors.Is/As: those walk the chain and would say
	// yes for a wrapped e too, whereas the question here is whether anything was wrapped around it
	// at all. errors.AsType above already established that e is somewhere in the chain.
	if err == error(e) { //nolint:err113,errorlint
		return e // already is an HPError with nothing wrapped around it
	}

	// The caller added context around an existing HPError, typically fmt.Errorf("...: %w", err).
	// Returning the inner HPError would throw that context away, so keep its identity - error
	// code, status, params and stack trace all stay reachable through the chain - while adopting
	// the outer error so the added message survives.
	//
	// NOTE: this has to be a copy. Assigning the outer error onto e.err would make e.Unwrap()
	// lead back to e, and every chain walker here would loop forever.
	wrapped := *e
	wrapped.params = cloneErrParams(e.params)
	wrapped.ntParams = cloneErrParams(e.ntParams)
	wrapped.err = err
	return &wrapped
}

func cloneErrParams(params map[string]any) map[string]any {
	if params == nil {
		return map[string]any{}
	}
	return maps.Clone(params)
}
