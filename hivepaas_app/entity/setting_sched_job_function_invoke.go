package entity

// SchedJobFunctionInvoke is the request a function-invoke job calls its
// function with, as the runtime's invoke reads it: the query is in the path,
// the body is text.
type SchedJobFunctionInvoke struct {
	Method string `json:"method"`
	// Path is the path the handler sees, with its query: "/report?day=today".
	Path string `json:"path"`
	// Headers are by name, in lower case.
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
}
