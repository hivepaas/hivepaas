package backupsnapshotdto

import (
	"path"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const maxRestorePathLen = 500

// RestoreBackupSnapshotReq puts a snapshot back into an app: a command
// snapshot's file through Command, a volume snapshot into a volume the app
// mounts.
type RestoreBackupSnapshotReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
	ID    string              `json:"-" mapstructure:"-"`

	TargetApp basedto.ObjectIDReq `json:"targetApp"`
	// Command loads the file of a command snapshot from its stdin.
	Command *commandtemplatedto.CommandTemplateBaseReq `json:"command"`
	// Volume is a volume the app mounts as its own directory; Subpath a path
	// inside what the app sees of it, where the snapshot's root goes.
	Volume  basedto.ObjectIDReq `json:"volume"`
	Subpath string              `json:"subpath"`
	// SnapshotPath is a directory inside the snapshot to restore alone; "" for
	// all of it.
	SnapshotPath string                 `json:"snapshotPath"`
	StopApp      bool                   `json:"stopApp"`
	Mode         base.BackupRestoreMode `json:"mode"`
}

func NewRestoreBackupSnapshotReq() *RestoreBackupSnapshotReq {
	return &RestoreBackupSnapshotReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *RestoreBackupSnapshotReq) ModifyRequest() error {
	req.Subpath = cleanRelativePath(req.Subpath)
	req.SnapshotPath = cleanRelativePath(req.SnapshotPath)
	if req.Command != nil {
		req.Command.Name = "-"
		req.Command.Kind = ""
		// It reads the file on its stdin, which a TTY would not pass through as is.
		req.Command.TTY = false
		if err := req.Command.ModifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *RestoreBackupSnapshotReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	validators = append(validators, basedto.ValidateObjectIDReq(&req.TargetApp, true, "targetApp")...)

	isCommand := req.Command != nil
	isVolume := req.Volume.ID != ""
	validators = append(validators, basedto.ValidateCond(isCommand != isVolume, "command")...)
	if isCommand {
		validators = append(validators, req.Command.Validate("command")...)
	}
	validators = append(validators, basedto.ValidateObjectIDReq(&req.Volume, false, "volume")...)
	validators = append(validators, basedto.ValidateCond(isVolume || req.Subpath == "", "subpath")...)
	validators = append(validators, basedto.ValidateCond(isRelativePath(req.Subpath), "subpath")...)
	validators = append(validators, basedto.ValidateCond(isVolume || req.SnapshotPath == "", "snapshotPath")...)
	validators = append(validators, basedto.ValidateCond(isRelativePath(req.SnapshotPath), "snapshotPath")...)
	validators = append(validators, basedto.ValidateCond(isVolume || !req.StopApp, "stopApp")...)
	if isVolume {
		validators = append(validators, basedto.ValidateStrIn(&req.Mode, true, base.AllBackupRestoreModes,
			"mode")...)
		// Its containers would go on writing to the directory moved aside.
		validators = append(validators, basedto.ValidateCond(
			req.Mode != base.BackupRestoreModeReplace || req.StopApp, "stopApp")...)
	} else {
		validators = append(validators, basedto.ValidateCond(req.Mode == "", "mode")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type RestoreBackupSnapshotResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *RestoreBackupSnapshotDataResp `json:"data"`
}

type RestoreBackupSnapshotDataResp struct {
	// Task is the restore's.
	Task *basedto.ObjectIDResp `json:"task"`
}

// cleanRelativePath tidies a relative path; one from the root is left for the
// check to refuse.
func cleanRelativePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || strings.HasPrefix(p, "/") {
		return p
	}
	p = strings.TrimSuffix(path.Clean(p), "/")
	if p == "." {
		return ""
	}
	return p
}

// isRelativePath is a path inside a directory: not from the root, never above
// where it starts.
func isRelativePath(p string) bool {
	if p == "" {
		return true
	}
	if len(p) > maxRestorePathLen || strings.HasPrefix(p, "/") {
		return false
	}
	cleaned := path.Clean(p)
	return cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}
