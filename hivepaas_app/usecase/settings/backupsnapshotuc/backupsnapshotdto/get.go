package backupsnapshotdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetBackupSnapshotReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-" mapstructure:"-"`
}

func NewGetBackupSnapshotReq() *GetBackupSnapshotReq {
	return &GetBackupSnapshotReq{}
}

func (req *GetBackupSnapshotReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ID, true, "id")...))
}

type GetBackupSnapshotResp struct {
	Meta *basedto.Meta       `json:"meta"`
	Data *BackupSnapshotResp `json:"data"`
}

type DeleteBackupSnapshotReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-" mapstructure:"-"`
}

func NewDeleteBackupSnapshotReq() *DeleteBackupSnapshotReq {
	return &DeleteBackupSnapshotReq{}
}

func (req *DeleteBackupSnapshotReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateID(&req.ID, true, "id")...))
}

type DeleteBackupSnapshotResp struct {
	Meta *basedto.Meta `json:"meta"`
}
