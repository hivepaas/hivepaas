package appmanageddto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// AddManagedReplicaReq makes an asynchronous read replica of a managed database,
// on a volume of its own: pinned to a node, preferably not the database's.
type AddManagedReplicaReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	DataVolume *basedto.ObjectIDReq `json:"dataVolume"`
}

func NewAddManagedReplicaReq() *AddManagedReplicaReq {
	return &AddManagedReplicaReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *AddManagedReplicaReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateObjectIDReq(req.DataVolume, true,
		"dataVolume")...))
}

type AddManagedReplicaResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *AddManagedReplicaDataResp `json:"data"`
}

// AddManagedReplicaDataResp is the replica's app, and the task of its first
// deployment, which seeds it from the archive.
type AddManagedReplicaDataResp struct {
	App  *basedto.ObjectIDResp `json:"app"`
	Task *basedto.ObjectIDResp `json:"task"`
}

type RemoveManagedReplicaReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	MemberID     string `json:"-"`
}

func NewRemoveManagedReplicaReq() *RemoveManagedReplicaReq {
	return &RemoveManagedReplicaReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *RemoveManagedReplicaReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.MemberID, true, "memberId")...))
}

type RemoveManagedReplicaResp struct {
	Meta *basedto.Meta `json:"meta"`
}
