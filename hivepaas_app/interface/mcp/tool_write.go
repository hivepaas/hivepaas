package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// applier is how apply_plan carries out the plans one tool makes.
type applier struct {
	// check compares what the plan saw with what is there now, and refuses a
	// plan that no longer holds; nil for a change that depends on nothing.
	check func(ctx context.Context, call *Call, seen json.RawMessage) error
	// result trims the endpoint's data to what a model needs; nil keeps it.
	result func(data json.RawMessage) (any, error)
	// refused words the endpoint's refusal of a plan that no longer holds; nil,
	// or a nil answer, leaves the refusal as the endpoint worded it.
	refused func(apiErr *APIError) error
	// follow says which read tools watch the change happen.
	follow string
	// needs is what the plan tool, and so its apply, needs of the caller.
	needs Need
}

// planResult is what every plan tool answers: its plan, and the token that
// applies it when there is one.
type planResult[T any] struct {
	Plan      T         `json:"plan"`
	PlanToken string    `json:"planToken,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitzero"`
	Next      string    `json:"next"`
}

const (
	nextApply = "Show the person this plan in full. Call apply_plan with planToken only once they " +
		"have agreed to it; if they want it changed, plan again."
	nextNothing = "There is nothing to apply: the plan says what stops it, or that nothing would change."
)

// planTool is a tool that changes nothing and answers a plan. run answers what
// it would do and, when it can be done, the request that does it; the plan is
// then kept for apply_plan, and the answer carries its token. A plan's audit
// entry names the plan, so the apply's entry can be traced to it.
func planTool[In, T any](name, title, description string, needs Need, applies *applier,
	run func(ctx context.Context, call *Call, in In) (T, *storedPlan, error)) Tool {
	applies.needs = needs
	return Tool{Name: name, Title: title, Description: description, Kind: KindPlan, needs: needs, applies: applies,
		add: func(s *mcpsdk.Server, deps *Deps) {
			sdkTool := &mcpsdk.Tool{Name: name, Title: title, Description: description,
				Annotations: &mcpsdk.ToolAnnotations{ReadOnlyHint: true, Title: title}}
			addTool(s, deps, sdkTool, false,
				func(ctx context.Context, call *Call, c *caller, in In) (planResult[T], error) {
					answer, plan, err := run(ctx, call, in)
					if err != nil {
						if recErr := recordCall(ctx, deps, c, name, in, base.AuditLogResultAllowed); recErr != nil {
							return planResult[T]{}, recErr
						}
						return planResult[T]{}, err
					}
					out := planResult[T]{Plan: answer, Next: nextNothing}
					if plan == nil {
						return out, recordCall(ctx, deps, c, name, in, base.AuditLogResultAllowed)
					}
					// The answer must fit before a plan is kept for it.
					if err = boundAnswer(&out); err != nil {
						return planResult[T]{}, err
					}
					plan.Tool = name
					token, planID, err := savePlan(ctx, deps.Plans, c, plan)
					if err != nil {
						return planResult[T]{}, err
					}
					out.PlanToken, out.ExpiresAt, out.Next = token, timeNow().Add(planTTL).UTC(), nextApply
					return out, recordCall(ctx, deps, c, name, in, base.AuditLogResultAllowed,
						auditNote{noteKeyPlan, planID}, auditNote{"summary", plan.Summary})
				})
		}}
}

// noteKeyPlan is the audit note that names a plan, on its entry and its apply's.
const noteKeyPlan = "plan"

type applyInput struct {
	PlanToken string `json:"planToken" jsonschema:"the planToken a plan_* tool answered"`
}

// forAudit keeps the token out of the log: the plan's ID, noted beside it, says
// which plan it was.
func (in applyInput) forAudit() any {
	return applyInput{PlanToken: "(given)"}
}

type applyAnswer struct {
	Tool    string `json:"tool"`
	Summary string `json:"summary"`
	Result  any    `json:"result,omitempty"`
	Follow  string `json:"follow,omitempty"`
}

var errChangesNotAllowed = &InputError{Message: "changes are not allowed: an administrator allows them in " +
	"the dashboard, System settings, AI, MCP server"}

func applyPlanTool() Tool {
	const name, title = "apply_plan", "Apply a plan"
	const description = "Carries out a plan a plan_* tool made, exactly as it was shown. Call it only once " +
		"the person has seen the plan and agreed to it. A plan is used once and lasts ten minutes."
	destructive, openWorld := true, false
	return Tool{Name: name, Title: title, Description: description, Kind: KindApply, needs: NeedChange,
		add: func(s *mcpsdk.Server, deps *Deps) {
			sdkTool := &mcpsdk.Tool{Name: name, Title: title, Description: description,
				Annotations: &mcpsdk.ToolAnnotations{Title: title, DestructiveHint: &destructive,
					OpenWorldHint: &openWorld}}
			addTool(s, deps, sdkTool, false, applyPlan)
		}}
}

func applyPlan(ctx context.Context, call *Call, c *caller, in applyInput) (applyAnswer, error) {
	deps := call.deps
	if refusal := c.access.refusal(NeedChange); refusal != nil {
		if recErr := recordCall(ctx, deps, c, "apply_plan", in, base.AuditLogResultDenied); recErr != nil {
			return applyAnswer{}, recErr
		}
		return applyAnswer{}, refusal
	}
	plan, planID, err := takePlan(ctx, deps.Plans, c, in.PlanToken)
	var applies *applier
	if err == nil {
		if applies = deps.appliers[plan.Tool]; applies == nil {
			err = errNoSuchPlan
		}
	}
	// The plan's own kind of change: a key that may restart apps cannot apply a
	// plan to install one, whoever made it.
	if err == nil {
		if refusal := c.access.refusal(applies.needs); refusal != nil {
			if recErr := recordCall(ctx, deps, c, "apply_plan", in, base.AuditLogResultDenied,
				auditNote{noteKeyPlan, planID}, auditNote{"applies", plan.Tool}); recErr != nil {
				return applyAnswer{}, recErr
			}
			return applyAnswer{}, refusal
		}
	}
	if err != nil {
		if recErr := recordCall(ctx, deps, c, "apply_plan", in, base.AuditLogResultAllowed); recErr != nil {
			return applyAnswer{}, recErr
		}
		return applyAnswer{}, err
	}
	// Recorded before anything is sent: a change whose record cannot be written
	// is not made. Named with the tool it applies, which is what a list of calls
	// shows of an entry: its detail is cut there.
	if err = recordCall(ctx, deps, c, "apply_plan: "+plan.Tool, in, base.AuditLogResultAllowed,
		auditNote{noteKeyPlan, planID}, auditNote{"applies", plan.Tool}, auditNote{"summary", plan.Summary}); err != nil {
		return applyAnswer{}, err
	}
	if applies.check != nil && len(plan.Check) > 0 {
		if err = applies.check(ctx, call, plan.Check); err != nil {
			return applyAnswer{}, err
		}
	}
	return sendPlan(ctx, call, plan, applies)
}

// sendPlan sends a plan's request, as the caller, and reads what it did.
func sendPlan(ctx context.Context, call *Call, plan *storedPlan, applies *applier) (applyAnswer, error) {
	resp, err := call.Do(ctx, plan.Method, plan.Path, nil, plan.Body)
	if err != nil {
		var apiErr *APIError
		if applies.refused != nil && errors.As(err, &apiErr) {
			if worded := applies.refused(apiErr); worded != nil {
				return applyAnswer{}, worded
			}
		}
		return applyAnswer{}, err
	}
	out := applyAnswer{Tool: plan.Tool, Summary: plan.Summary, Follow: applies.follow}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if len(resp.Body) > 0 {
		if err = json.Unmarshal(resp.Body, &envelope); err != nil {
			return applyAnswer{}, fmt.Errorf("mcp: decoding the answer: %w", err)
		}
	}
	if len(envelope.Data) > 0 {
		if applies.result != nil {
			out.Result, err = applies.result(envelope.Data)
		} else {
			err = json.Unmarshal(envelope.Data, &out.Result)
		}
		if err != nil {
			return applyAnswer{}, fmt.Errorf("mcp: decoding the answer: %w", err)
		}
	}
	return out, nil
}

// planMoved is the refusal of a plan whose state changed since it was made.
func planMoved(what string) error {
	return &InputError{Message: what + " changed since the plan was made; nothing was applied. Plan again."}
}
