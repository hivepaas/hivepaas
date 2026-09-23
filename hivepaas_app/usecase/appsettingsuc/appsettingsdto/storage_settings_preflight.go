package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// PreflightAppStorageSettingsReq carries the mounts a screen is about to save.
//
// It is the update's body without its version: nothing is written, and refusing
// a question because somebody else saved in the meantime would only make the
// warning disappear when it is most needed.
type PreflightAppStorageSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Mounts []*Mount `json:"mounts"`
}

func NewPreflightAppStorageSettingsReq() *PreflightAppStorageSettingsReq {
	return &PreflightAppStorageSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *PreflightAppStorageSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type PreflightAppStorageSettingsResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data *PreflightAppStorageResult `json:"data"`
}

type PreflightAppStorageResult struct {
	// Storage is every mount being added whose directory already holds
	// something. Empty is the ordinary case.
	Storage []*PreflightStorageRes `json:"storage"`
}

type PreflightStorageRes struct {
	// Target is the path inside the container, which is how the screen finds the
	// row this is about.
	Target string `json:"target"`
	Volume struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"volume"`
	// Path is the directory inside the volume, so an operator can go and look.
	Path string `json:"path"`
}
