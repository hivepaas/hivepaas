package hpappsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const changeIDMaxLen = 40

// PendingChangeResp is what a caller needs to finish a routing change: which
// change to confirm, when confirming starts meaning something, and when the
// chance is gone.
//
// It is returned by the update, and repeated on the GET, so a dashboard that was
// reloaded mid-trial can pick the countdown back up instead of leaving the change
// to expire because the tab was refreshed.
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

type ConfirmRoutingSettingsReq struct {
	// ChangeID names the change being vouched for. It is optional, but sending it
	// is what stops a confirmation that was in flight during one change from
	// landing on the next one.
	ChangeID string `json:"changeId"`
}

func NewConfirmRoutingSettingsReq() *ConfirmRoutingSettingsReq {
	return &ConfirmRoutingSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ConfirmRoutingSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ChangeID, true, 0, changeIDMaxLen, "changeId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ConfirmRoutingSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type RevertRoutingSettingsReq struct {
	ChangeID string `json:"changeId"`
}

func NewRevertRoutingSettingsReq() *RevertRoutingSettingsReq {
	return &RevertRoutingSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RevertRoutingSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ChangeID, true, 0, changeIDMaxLen, "changeId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RevertRoutingSettingsResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *RevertRoutingDataResp `json:"data"`
}

type RevertRoutingDataResp struct {
	Reverted bool   `json:"reverted"`
	Reason   string `json:"reason,omitempty"`
}
