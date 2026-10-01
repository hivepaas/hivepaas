package entity

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

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

// SchedJobFunctionInvokeResult is what a function's call answered: its task's
// output.
type SchedJobFunctionInvokeResult struct {
	// Outcome is the runtime's: ok, error or timeout.
	Outcome string              `json:"outcome"`
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers,omitempty"`
	// Body is in base64; BodyTruncated says it was cut.
	Body          []byte  `json:"body,omitempty"`
	BodyTruncated bool    `json:"bodyTruncated,omitempty"`
	RequestID     string  `json:"requestId,omitempty"`
	DurationMs    float64 `json:"durationMs"`
}

// OutputAsFunctionInvoke is the response a function's call got; nil for a task
// that holds none, another job's among them.
func (t *Task) OutputAsFunctionInvoke() (*SchedJobFunctionInvokeResult, error) {
	if t.Output == "" {
		return nil, nil //nolint:nilnil // no output, no response
	}
	// Read apart from the task's parsed output, as a data backup's is.
	result := &SchedJobFunctionInvokeResult{}
	if err := json.Unmarshal([]byte(t.Output), result); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if result.Outcome == "" {
		return nil, nil //nolint:nilnil // no outcome: another job's output
	}
	return result, nil
}
