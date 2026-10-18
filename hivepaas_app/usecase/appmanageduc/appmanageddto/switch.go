package appmanageddto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const maxSwitchConsumers = 200

type GetManagedSwitchReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetManagedSwitchReq() *GetManagedSwitchReq {
	return &GetManagedSwitchReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetManagedSwitchReq) Validate() hperrors.ValidationErrors {
	return validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID)
}

type GetManagedSwitchResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *ManagedSwitchDataResp `json:"data"`
}

// ManagedSwitchDataResp is what a switch to a restored database would do: the
// database it was restored from, which keeps running, and the apps of the
// environment that use it.
type ManagedSwitchDataResp struct {
	Source *ManagedSwitchAppResp `json:"source"`
	// Consumers are the apps that name the source: by a variable's reference to
	// its key, by its id among their previews' databases to clone, or as the
	// database their deployments mark a restore point on.
	Consumers []*ManagedSwitchConsumerResp `json:"consumers"`
	// SwitchedAt is when consumers were last switched to this database, null
	// before the first time.
	SwitchedAt *time.Time `json:"switchedAt"`
}

type ManagedSwitchAppResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ManagedSwitchConsumerResp struct {
	App *ManagedSwitchAppResp `json:"app"`
	// EnvVars are the app's variables that name the source's.
	EnvVars []string `json:"envVars"`
	// PreviewClone says the app's previews clone the source.
	PreviewClone bool `json:"previewClone"`
	// DeployRestorePoint says the app's deployments mark a restore point on the
	// source.
	DeployRestorePoint bool `json:"deployRestorePoint"`
}

// SwitchManagedReq switches the consumers chosen from the database this one was
// restored from to this one, and pauses the source's jobs. Their variables are
// applied to their running services, as saving them does; with redeploy, each
// whose variables changed is deployed too, for what reads the database as it
// builds or in its pre-deploy command.
type SwitchManagedReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Consumers []*basedto.ObjectIDReq `json:"consumers"`
	Redeploy  bool                   `json:"redeploy"`
}

func NewSwitchManagedReq() *SwitchManagedReq {
	return &SwitchManagedReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SwitchManagedReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	validators := make([]vld.Validator, 0, len(req.Consumers)+1)
	validators = append(validators, basedto.ValidateCond(len(req.Consumers) <= maxSwitchConsumers, "consumers")...)
	for _, consumer := range req.Consumers {
		validators = append(validators, basedto.ValidateObjectIDReq(consumer, true, "consumers")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type SwitchManagedResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *SwitchManagedDataResp `json:"data"`
}

type SwitchManagedDataResp struct {
	// Switched are the consumers switched.
	Switched []*basedto.ObjectIDResp `json:"switched"`
	// Deployments are the consumers' deployments queued, with redeploy.
	Deployments []*ManagedSwitchDeploymentResp `json:"deployments"`
}

type ManagedSwitchDeploymentResp struct {
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
	Task       *basedto.ObjectIDResp `json:"task"`
}
