package appmanageddto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedscripts"
)

type GetManagedScriptsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetManagedScriptsReq() *GetManagedScriptsReq {
	return &GetManagedScriptsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetManagedScriptsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetManagedScriptsResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *ManagedScriptsResp `json:"data"`
}

// ManagedScriptsResp is what the installation's revision of a managed database's
// template offers it of what its engine owns - its scripts and artifacts - and
// whether it can be applied.
type ManagedScriptsResp struct {
	// State is up-to-date, update-available, incompatible or unavailable.
	State string `json:"state"`
	// Reason says why the revision is incompatible or unavailable.
	Reason string `json:"reason"`
	// Current is what the database runs: the revision its scripts came from, empty
	// when that is not known, and their fingerprint.
	Current *ManagedScriptsRevisionResp `json:"current"`
	// Available is the installation's revision of the template and the fingerprint
	// of what it offers: null when it offers nothing.
	Available *ManagedScriptsRevisionResp `json:"available"`
	// Changes are what applying it changes: each script by its path, then each
	// artifact by its name.
	Changes []*ManagedScriptChangeResp `json:"changes"`
	// OutOfStep are the database's members and verifier whose scripts differ from
	// the database's.
	OutOfStep []*ManagedScriptsMemberResp `json:"outOfStep"`
}

type ManagedScriptsRevisionResp struct {
	Revision    string `json:"revision"`
	Fingerprint string `json:"fingerprint"`
}

type ManagedScriptChangeResp struct {
	// Kind is script or artifact.
	Kind string `json:"kind"`
	// Name is a script's path, or an artifact's name.
	Name string `json:"name"`
	// Change is added, changed or removed.
	Change string `json:"change"`
	// From and To are an artifact's versions.
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

type ManagedScriptsMemberResp struct {
	App *basedto.ObjectIDResp `json:"app"`
	Key string                `json:"key"`
}

// ManagedScriptsView is what the use case read of a database's scripts.
type ManagedScriptsView struct {
	Offer             *managedscripts.Offer
	Current           *managedscripts.Set
	CurrentRevision   string
	AvailableRevision string
	OutOfStep         []*ManagedScriptsMemberResp
}

func TransformManagedScripts(view *ManagedScriptsView) *ManagedScriptsResp {
	resp := &ManagedScriptsResp{
		State:     view.Offer.State,
		Reason:    view.Offer.Reason,
		Current:   &ManagedScriptsRevisionResp{Revision: view.CurrentRevision, Fingerprint: view.Current.Fingerprint()},
		Changes:   make([]*ManagedScriptChangeResp, 0, len(view.Offer.Changes)),
		OutOfStep: view.OutOfStep,
	}
	if resp.OutOfStep == nil {
		resp.OutOfStep = []*ManagedScriptsMemberResp{}
	}
	if view.Offer.Available != nil {
		resp.Available = &ManagedScriptsRevisionResp{Revision: view.AvailableRevision,
			Fingerprint: view.Offer.Available.Fingerprint()}
	}
	for _, change := range view.Offer.Changes {
		resp.Changes = append(resp.Changes, &ManagedScriptChangeResp{Kind: change.Kind, Name: change.Name,
			Change: change.Change, From: change.From, To: change.To})
	}
	return resp
}

type UpgradeManagedScriptsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	// Fingerprint is the available fingerprint the person reviewed: the upgrade is
	// refused when the revision offers another since.
	Fingerprint string `json:"fingerprint"`
}

func NewUpgradeManagedScriptsReq() *UpgradeManagedScriptsReq {
	return &UpgradeManagedScriptsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpgradeManagedScriptsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStr(&req.Fingerprint, true, fingerprintLen, fingerprintLen,
		"fingerprint")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// fingerprintLen is a fingerprint's length: a sha256, in hex.
const fingerprintLen = 64

type UpgradeManagedScriptsResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *UpgradeManagedScriptsDataResp `json:"data"`
}

// UpgradeManagedScriptsDataResp is the revision applied, and the deployments
// that bring it to the database, its members and verifier: the member that
// writes last.
type UpgradeManagedScriptsDataResp struct {
	Revision    string                   `json:"revision"`
	Fingerprint string                   `json:"fingerprint"`
	Deployments []*ManagedDeploymentResp `json:"deployments"`
}

type ManagedDeploymentResp struct {
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
	Task       *basedto.ObjectIDResp `json:"task"`
}
