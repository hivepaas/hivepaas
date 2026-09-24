package specdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// ValidateImportReq asks what importing a bundle at a scope would do.
//
// The bundle travels inside the JSON body, base64-encoded, with every call:
// nothing is kept on the server between validate and apply, so a bundle that
// carries secrets is never stored.
type ValidateImportReq struct {
	Scope *entity.ObjectScope `json:"-"`

	// Bundle is the archive export produced, as it was downloaded. JSON carries
	// it as base64.
	Bundle []byte `json:"bundle"`
	// Passphrase opens an encrypted bundle. It is never stored or logged.
	Passphrase string                  `json:"passphrase"`
	Selection  specmodel.Selection     `json:"selection"`
	Options    specmodel.ImportOptions `json:"options"`
}

func NewValidateImportReq() *ValidateImportReq {
	return &ValidateImportReq{}
}

// ModifyRequest defaults the option most people mean: make what exists match.
func (req *ValidateImportReq) ModifyRequest() error {
	if req.Options.Existing == "" {
		req.Options.Existing = specmodel.ExistingUpdate
	}
	return nil
}

var _ basedto.ReqModifier = (*ValidateImportReq)(nil)

func (req *ValidateImportReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateSlice(req.Bundle, false, 1, nil, "bundle")...)
	validators = append(validators, basedto.ValidateStr(&req.Passphrase, false,
		1, passphraseMaxLen, "passphrase")...)
	validators = append(validators, basedto.ValidateStrIn(&req.Options.Existing, true,
		[]specmodel.Existing{specmodel.ExistingUpdate, specmodel.ExistingKeep}, "options.existing")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ApplyImportReq applies what a validate of the same body planned. PlanHash is
// the plan the operator saw: a plan that changed since is refused, and shown
// again.
type ApplyImportReq struct {
	ValidateImportReq
	PlanHash string `json:"planHash"`
	// AcceptIssues accepts every skipped, fixable and warning issue of the plan.
	// A plan with any is refused without it.
	AcceptIssues bool `json:"acceptIssues"`
}

func NewApplyImportReq() *ApplyImportReq {
	return &ApplyImportReq{}
}

func (req *ApplyImportReq) Validate() hperrors.ValidationErrors {
	errs := req.ValidateImportReq.Validate()
	validators := basedto.ValidateStr(&req.PlanHash, true, 1, planHashMaxLen, "planHash")
	return append(errs, hperrors.NewValidationErrors(vld.Validate(validators...))...)
}

type ApplyImportResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *ApplyImportData `json:"data"`
}

type ApplyImportData struct {
	// Plan is the plan applied, each selected node with its outcome.
	Plan *specmodel.ImportPlan `json:"plan"`
	// Deployments are the deployments the import queued.
	Deployments []*specservice.ImportDeployment `json:"deployments"`
}

const planHashMaxLen = 128

type ValidateImportResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *specmodel.ImportPlan `json:"data"`
}
