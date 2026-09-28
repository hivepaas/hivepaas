package backupsnapshotdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// ListBackupSnapshotEntriesReq lists a directory of a snapshot.
type ListBackupSnapshotEntriesReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-" mapstructure:"-"`
	// Path is a directory inside the snapshot; "" for its root.
	Path string `json:"-" mapstructure:"path"`
}

func NewListBackupSnapshotEntriesReq() *ListBackupSnapshotEntriesReq {
	return &ListBackupSnapshotEntriesReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *ListBackupSnapshotEntriesReq) ModifyRequest() error {
	req.Path = cleanRelativePath(req.Path)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *ListBackupSnapshotEntriesReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateCond(isRelativePath(req.Path), "path")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListBackupSnapshotEntriesResp struct {
	Meta *basedto.Meta              `json:"meta"`
	Data []*BackupSnapshotEntryResp `json:"data"`
}

// BackupSnapshotEntryResp is a file or a directory a snapshot holds.
type BackupSnapshotEntryResp struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir,omitempty"`
	// SizeBytes is a directory's too: all it holds.
	SizeBytes int64 `json:"sizeBytes"`
}
