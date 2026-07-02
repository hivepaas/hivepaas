package apperrors

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

	InnerErrors []*InnerErrorInfo `json:"errors,omitempty"`
}

type InnerErrorInfo struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Cause   string `json:"cause,omitempty"`
}
