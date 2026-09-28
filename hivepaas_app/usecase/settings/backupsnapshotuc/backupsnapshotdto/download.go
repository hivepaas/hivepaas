package backupsnapshotdto

import (
	"context"
	"io"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// DownloadBackupSnapshotFileReq asks for a file of a snapshot.
type DownloadBackupSnapshotFileReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-" mapstructure:"-"`
	// Path is a file inside the snapshot.
	Path string `json:"-" mapstructure:"path"`
}

func NewDownloadBackupSnapshotFileReq() *DownloadBackupSnapshotFileReq {
	return &DownloadBackupSnapshotFileReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *DownloadBackupSnapshotFileReq) ModifyRequest() error {
	req.Path = cleanRelativePath(req.Path)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *DownloadBackupSnapshotFileReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateCond(req.Path != "" && isRelativePath(req.Path), "path")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// DownloadBackupSnapshotFileResp is the file, to be written to the response once
// its headers are.
type DownloadBackupSnapshotFileResp struct {
	FileName  string
	SizeBytes int64
	// Write writes the file's content to w. It fails half way when the engine
	// does; the caller has sent the headers by then.
	Write func(ctx context.Context, w io.Writer) error
}
