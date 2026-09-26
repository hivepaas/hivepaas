package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

// Kind is what a tool may do. A read tool is served to every caller; a plan
// and the apply only to one whose request may change things.
type Kind string

const (
	KindRead Kind = "read"
	// KindPlan changes nothing, and answers a plan that apply_plan can carry out.
	KindPlan Kind = "plan"
	// KindApply carries out a plan.
	KindApply Kind = "apply"
)

// changes is whether a tool is served only to a caller who may change things.
func (k Kind) changes() bool {
	return k != KindRead
}

// Deps is what every tool is served with.
type Deps struct {
	Dispatcher *Dispatcher
	Audit      auditservice.Service
	DB         database.IDB
	Settings   settingsReader
	Plans      planRepo
	// appliers are the plan tools' appliers by tool name, for apply_plan.
	appliers map[string]*applier
}

// Tool is one tool, as it is registered with the SDK's server.
type Tool struct {
	Name        string
	Title       string
	Description string
	Kind        Kind
	add         func(s *mcpsdk.Server, deps *Deps)
	// applies is how apply_plan carries out a plan this tool made; nil for a
	// tool that makes none.
	applies *applier
}

// InputError is a request a tool cannot answer as asked - an ambiguous name, a
// value out of range. It becomes a tool error the model reads and can correct.
type InputError struct {
	Message string
}

func (e *InputError) Error() string { return e.Message }

// Call is what a tool's run function works with.
type Call struct {
	deps *Deps
}

// Do dispatches a request as the caller and returns the handler's answer.
func (c *Call) Do(ctx context.Context, method, path string, query url.Values, body any) (*Response, error) {
	return c.deps.Dispatcher.Do(ctx, method, path, query, body)
}

// Get dispatches a GET and decodes the answer into out.
func (c *Call) Get(ctx context.Context, path string, query url.Values, out any) error {
	return c.decode(c.Do(ctx, http.MethodGet, path, query, nil))(out)
}

// Post dispatches a POST with a JSON body and decodes the answer into out.
func (c *Call) Post(ctx context.Context, path string, body, out any) error {
	return c.decode(c.Do(ctx, http.MethodPost, path, nil, body))(out)
}

func (c *Call) decode(resp *Response, err error) func(out any) error {
	return func(out any) error {
		if err != nil {
			return err
		}
		if out == nil {
			return nil
		}
		if err := json.Unmarshal(resp.Body, out); err != nil {
			return fmt.Errorf("mcp: decoding the answer: %w", err)
		}
		return nil
	}
}

// readTool is a tool that changes nothing. Its input schema is In's and its
// answer Out's, both from their JSON and jsonschema tags. Every call is audited
// before it runs, an answer is bounded to MaxToolOutput, and an error a model
// can act on - not found, not permitted, a bad input - becomes a tool error
// rather than a protocol one.
func readTool[In, Out any](name, title, description string,
	run func(ctx context.Context, call *Call, in In) (Out, error)) Tool {
	return Tool{Name: name, Title: title, Description: description, Kind: KindRead,
		add: func(s *mcpsdk.Server, deps *Deps) {
			sdkTool := &mcpsdk.Tool{Name: name, Title: title, Description: description,
				Annotations: &mcpsdk.ToolAnnotations{ReadOnlyHint: true, Title: title}}
			addTool(s, deps, sdkTool, true, func(ctx context.Context, call *Call, _ *caller, in In) (Out, error) {
				return run(ctx, call, in)
			})
		}}
}

// addTool registers a tool's handler with what every tool does around it. With
// auditFirst the call is recorded before it runs; otherwise run records it, when
// it knows what the record should say - which plan, which change.
func addTool[In, Out any](s *mcpsdk.Server, deps *Deps, sdkTool *mcpsdk.Tool, auditFirst bool,
	run func(ctx context.Context, call *Call, c *caller, in In) (Out, error)) {
	name := sdkTool.Name
	mcpsdk.AddTool(s, sdkTool, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (
		*mcpsdk.CallToolResult, Out, error) {
		var zero Out
		c := callerFrom(ctx)
		if c == nil {
			return nil, zero, errNoCaller
		}
		if auditFirst {
			if err := recordCall(ctx, deps, c, name, in, base.AuditLogResultAllowed); err != nil {
				return nil, zero, err
			}
		}

		out, err := run(ctx, &Call{deps: deps}, c, in)
		if err == nil {
			err = boundAnswer(any(&out))
		}
		if err == nil {
			return nil, out, nil
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) && (apiErr.Status == http.StatusUnauthorized ||
			apiErr.Status == http.StatusForbidden) {
			if recErr := recordCall(ctx, deps, c, name, in, base.AuditLogResultDenied); recErr != nil {
				return nil, zero, recErr
			}
		}
		if result := toolError(err); result != nil {
			return result, zero, nil
		}
		return nil, zero, err
	})
}

// toolError is the tool result for an error a model can act on, or nil for one
// it cannot, which the SDK answers as a protocol error.
func toolError(err error) *mcpsdk.CallToolResult {
	var apiErr *APIError
	var inputErr *InputError
	var text string
	switch {
	case errors.As(err, &inputErr):
		text = inputErr.Message
	case errors.As(err, &apiErr) && apiErr.Status < http.StatusInternalServerError:
		text = describeAPIError(apiErr)
	default:
		return nil
	}
	return &mcpsdk.CallToolResult{IsError: true, Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: text}}}
}

func describeAPIError(e *APIError) string {
	switch e.Status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return "not permitted: the API key's user may not do this, or the key is limited to less. " + e.Error()
	case http.StatusNotFound:
		return "not found, or not visible to the API key's user. " + e.Error()
	}
	return e.Error()
}

// auditedInput is an input that may carry a secret - a parameter an install
// asks for - and says what of it the audit log may keep.
type auditedInput interface {
	forAudit() any
}

// auditNote is something an entry says beside the input: the plan a call made
// or applied.
type auditNote struct {
	key   string
	value any
}

// recordCall writes the audit entry for one tool call. A call whose record
// cannot be written is not made, as for every recorded action.
func recordCall(ctx context.Context, deps *Deps, c *caller, tool string, input any,
	result base.AuditLogResult, notes ...auditNote) error {
	if in, ok := input.(auditedInput); ok {
		input = in.forAudit()
	}
	detail := auditdetail.New().Set("tool", tool).Set("input", input)
	for _, note := range notes {
		detail = detail.Set(note.key, note.value)
	}
	entry := &auditservice.Entry{
		Scope:   entity.NewObjectScopeGlobal().ScopeType,
		Type:    base.AuditLogTypeMCPToolCall,
		Source:  base.AuditLogSourceMCP,
		Result:  result,
		Auth:    c.auth,
		ResType: base.ResourceTypeMCP,
		ResName: tool,
		Detail:  detail.String(),
	}
	if err := deps.Audit.Record(ctx, deps.DB, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
