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

// Kind is what a tool may do. Phase 1 registers read tools only.
type Kind string

const (
	KindRead        Kind = "read"
	KindWrite       Kind = "write"
	KindDestructive Kind = "destructive"
)

// Deps is what every tool is served with.
type Deps struct {
	Dispatcher *Dispatcher
	Audit      auditservice.Service
	DB         database.IDB
}

// Tool is one tool, as it is registered with the SDK's server.
type Tool struct {
	Name        string
	Title       string
	Description string
	Kind        Kind
	add         func(s *mcpsdk.Server, deps *Deps)
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
			mcpsdk.AddTool(s, sdkTool, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (
				*mcpsdk.CallToolResult, Out, error) {
				var zero Out
				c := callerFrom(ctx)
				if c == nil {
					return nil, zero, errNoCaller
				}
				if err := recordCall(ctx, deps, c, name, in, base.AuditLogResultAllowed); err != nil {
					return nil, zero, err
				}

				out, err := run(ctx, &Call{deps: deps}, in)
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
		}}
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

// recordCall writes the audit entry for one tool call. A call whose record
// cannot be written is not made, as for every recorded action.
func recordCall(ctx context.Context, deps *Deps, c *caller, tool string, input any,
	result base.AuditLogResult) error {
	if in, ok := input.(auditedInput); ok {
		input = in.forAudit()
	}
	entry := &auditservice.Entry{
		Scope:   entity.NewObjectScopeGlobal().ScopeType,
		Type:    base.AuditLogTypeMCPToolCall,
		Source:  base.AuditLogSourceMCP,
		Result:  result,
		Auth:    c.auth,
		ResType: base.ResourceTypeMCP,
		ResName: tool,
		Detail:  auditdetail.New().Set("tool", tool).Set("input", input).String(),
	}
	if err := deps.Audit.Record(ctx, deps.DB, entry); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
