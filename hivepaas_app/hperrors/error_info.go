package hperrors

// ErrorInfo error info designed for JSON api.
type ErrorInfo struct {
	Type         string       `json:"type,omitempty"`
	Title        string       `json:"title"`
	Status       int          `json:"status"`
	Code         string       `json:"code"`
	Detail       string       `json:"detail"`
	Cause        string       `json:"cause,omitempty"`
	DebugLog     string       `json:"debugLog,omitempty"`
	StackTrace   string       `json:"stackTrace,omitempty"`
	DisplayLevel DisplayLevel `json:"displayLevel,omitempty"`
	// Params are values a client may act on, beside the message they appear in:
	// the limit a refused creation reached, for one. Most errors have none; which
	// an error has is part of what its code means.
	Params map[string]any `json:"params,omitempty"`

	InnerErrors []*InnerErrorInfo `json:"errors,omitempty"`
}

type InnerErrorInfo struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Cause   string `json:"cause,omitempty"`
}
