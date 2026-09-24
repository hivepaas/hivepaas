package appsettingsdto

import (
	"math"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ResetAppStoragePermissionsReq resets what is in the directory of one of an
// app's mounts, for an app that has to be given data another user wrote.
type ResetAppStoragePermissionsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	// Key is the mount's, as the storage settings list it.
	Key string `json:"key"`
	// Owner is who the directory and everything in it are given to. Without one,
	// every user is let read and write them instead.
	Owner *StorageOwnerReq `json:"owner"`
}

// StorageOwnerReq is a user and group by number. A name means something only
// inside the image that defines it.
type StorageOwnerReq struct {
	UID int `json:"uid"`
	GID int `json:"gid"`
}

func NewResetAppStoragePermissionsReq() *ResetAppStoragePermissionsReq {
	return &ResetAppStoragePermissionsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *ResetAppStoragePermissionsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateStr(&req.Key, true, 1, 100, "key")...) //nolint:mnd
	if req.Owner != nil {
		validators = append(validators, basedto.ValidateNumber(&req.Owner.UID, false, 0, math.MaxInt32, "owner.uid")...)
		validators = append(validators, basedto.ValidateNumber(&req.Owner.GID, false, 0, math.MaxInt32, "owner.gid")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ResetAppStoragePermissionsResp struct {
	Meta *basedto.Meta                       `json:"meta"`
	Data *ResetAppStoragePermissionsDataResp `json:"data"`
}

type ResetAppStoragePermissionsDataResp struct {
	// Path is the directory that was reset, inside its volume, so an operator can
	// go and look.
	Path string `json:"path"`
}
