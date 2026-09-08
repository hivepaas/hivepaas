package traefiksettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const changeIDMaxLen = 40

// PendingChangeResp is what a caller needs to finish a traefik config change:
// which change to confirm, when confirming starts meaning something, and when the
// chance is gone.
//
// It is returned by the update and repeated on the GET, and here the GET is the
// one that matters most. Applying these options replaces traefik's task, and
// traefik is the only way into HivePaaS - so the response to the update itself is
// sent down a connection that is about to be cut, and often never arrives. The
// dashboard has to be able to ask for the countdown again once it can reach the
// server, or a change nobody could confirm expires for want of a reply that was
// lost in transit.
//
// The twin of hpappsettingsdto.PendingChangeResp. Kept separate so each module's
// generated client gets its own named type, and so neither settings group's
// response shape is coupled to the other's.
type PendingChangeResp struct {
	ChangeID string `json:"changeId"`

	// ConfirmableFrom is when a confirmation starts being accepted. Before it,
	// the request would likely have arrived through the configuration being
	// replaced, so it proves nothing and is refused.
	ConfirmableFrom time.Time `json:"confirmableFrom"`
	DeadlineAt      time.Time `json:"deadlineAt"`
	AppliedAt       time.Time `json:"appliedAt"`
}

// TransformPendingChange describes a scheduled undo to its caller.
func TransformPendingChange(task *entity.Task) *PendingChangeResp {
	if task == nil {
		return nil
	}
	args, err := task.ArgsAsSettingsRevert()
	if err != nil || args == nil {
		return nil
	}
	return &PendingChangeResp{
		ChangeID:        task.ID,
		ConfirmableFrom: args.ConfirmableFrom(),
		DeadlineAt:      args.DeadlineAt,
		AppliedAt:       args.AppliedAt,
	}
}

type ConfirmConfigOptionsReq struct {
	// ChangeID names the change being vouched for. It is optional, but sending it
	// is what stops a confirmation that was in flight during one change from
	// landing on the next one.
	ChangeID string `json:"changeId"`
}

func NewConfirmConfigOptionsReq() *ConfirmConfigOptionsReq {
	return &ConfirmConfigOptionsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ConfirmConfigOptionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ChangeID, true, 0, changeIDMaxLen, "changeId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ConfirmConfigOptionsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type RevertConfigOptionsReq struct {
	ChangeID string `json:"changeId"`
}

func NewRevertConfigOptionsReq() *RevertConfigOptionsReq {
	return &RevertConfigOptionsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RevertConfigOptionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ChangeID, true, 0, changeIDMaxLen, "changeId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RevertConfigOptionsResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *RevertConfigOptionsData `json:"data"`
}

type RevertConfigOptionsData struct {
	Reverted bool   `json:"reverted"`
	Reason   string `json:"reason,omitempty"`
}
