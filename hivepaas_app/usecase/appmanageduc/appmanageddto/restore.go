package appmanageddto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	appNameMaxLen    = 100
	backupNameMaxLen = 64
)

// RestoreManagedReq restores a managed database into a new app, to one of four
// targets inside its restore window. The source app is only read.
type RestoreManagedReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Target ManagedRestoreTargetReq `json:"target"`
	// Name is the new app's: <app>-restored-<yyyymmdd-hhmmss> when empty.
	Name string `json:"name,omitempty"`
	// DataVolume is the cluster volume the new app keeps its data on, pinned to
	// a node the archive can be reached from.
	DataVolume *basedto.ObjectIDReq `json:"dataVolume"`
}

// ManagedRestoreTargetReq is what to recover to: latest, everything archived; a
// time; a restore point by name; or a base backup by name, as it was taken.
type ManagedRestoreTargetReq struct {
	Kind   string     `json:"kind"`
	Time   *time.Time `json:"time,omitempty"`
	Point  string     `json:"point,omitempty"`
	Backup string     `json:"backup,omitempty"`
}

func NewRestoreManagedReq() *RestoreManagedReq {
	return &RestoreManagedReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *RestoreManagedReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Target.Point = strings.TrimSpace(req.Target.Point)
	req.Target.Backup = strings.TrimSpace(req.Target.Backup)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *RestoreManagedReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	target := &req.Target
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateStrIn(&target.Kind, true, []string{entity.ManagedRestoreToLatest,
		entity.ManagedRestoreToTime, entity.ManagedRestoreToPoint, entity.ManagedRestoreToBackup}, "target.kind")...)
	validators = append(validators, basedto.ValidateCond((target.Kind == entity.ManagedRestoreToTime) ==
		(target.Time != nil), "target.time")...)
	validators = append(validators, basedto.ValidateCond((target.Kind == entity.ManagedRestoreToPoint) ==
		(target.Point != ""), "target.point")...)
	validators = append(validators, basedto.ValidateCond((target.Kind == entity.ManagedRestoreToBackup) ==
		(target.Backup != "") && len(target.Backup) <= backupNameMaxLen, "target.backup")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, false, 0, appNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateObjectIDReq(req.DataVolume, true, "dataVolume")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ValidatePoint checks the target point's name against the longest the
// source's engine takes, which is known once the source is read.
func (req *RestoreManagedReq) ValidatePoint(maxLen int) hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateCond(len(req.Target.Point) <= maxLen, "target.point")...))
}

type RestoreManagedResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *RestoreManagedDataResp `json:"data"`
}

type RestoreManagedDataResp struct {
	// App is the new app, restoring until the task says it is ready.
	App *basedto.ObjectIDResp `json:"app"`
	// Task sees the restore through: the recovery, the first base backup, and
	// what the app was restored with dropped.
	Task *basedto.ObjectIDResp `json:"task"`
}
