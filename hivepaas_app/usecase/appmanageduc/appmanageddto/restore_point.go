package appmanageddto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

const (
	pointNoteMaxLen = 500
	// MaxPointKeep is how far ahead a point may be kept: it holds back the
	// archive, every backup and log from it on, for as long.
	MaxPointKeep = 90 * 24 * time.Hour
)

// CreateRestorePointReq names a moment of a managed database, to restore to
// exactly. The name is a label - letters, digits, '.', '_' and '-', at most the
// engine's limit (63 for Postgres), not starting with "hp-" - and unique in the
// app.
type CreateRestorePointReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Name string `json:"name"`
	Note string `json:"note,omitempty"`
	// KeepUntil holds retention back for the point until then, at most 90 days
	// ahead; without it, the point lasts as long as retention keeps what it needs.
	KeepUntil *time.Time `json:"keepUntil,omitempty"`
}

func NewCreateRestorePointReq() *CreateRestorePointReq {
	return &CreateRestorePointReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *CreateRestorePointReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Note = strings.TrimSpace(req.Note)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateRestorePointReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	return hperrors.NewValidationErrors(vld.Validate(validatePoint(req.Note, req.KeepUntil)...))
}

// ValidateName checks the name against the longest the app's engine takes,
// which is known once the app is read.
func (req *CreateRestorePointReq) ValidateName(maxLen int) hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(basedto.ValidateStr(&req.Name, true, 1, maxLen, "name")...))
}

func validatePoint(note string, keepUntil *time.Time) []vld.Validator {
	now := timeutil.NowUTC()
	validators := basedto.ValidateStr(&note, false, 0, pointNoteMaxLen, "note")
	return append(validators, basedto.ValidateTime(keepUntil, false, now, now.Add(MaxPointKeep), "keepUntil")...)
}

type CreateRestorePointResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *ManagedRestorePointResp `json:"data"`
}

// UpdateRestorePointReq changes how long a point is kept, and its note: both as
// given, a null keepUntil keeping it no longer.
type UpdateRestorePointReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Name         string `json:"-"`

	KeepUntil *time.Time `json:"keepUntil"`
	Note      string     `json:"note"`
}

func NewUpdateRestorePointReq() *UpdateRestorePointReq {
	return &UpdateRestorePointReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *UpdateRestorePointReq) ModifyRequest() error {
	req.Note = strings.TrimSpace(req.Note)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateRestorePointReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	return hperrors.NewValidationErrors(vld.Validate(validatePoint(req.Note, req.KeepUntil)...))
}

type UpdateRestorePointResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// DeleteRestorePointReq removes a point's record. The mark stays in the
// database's log (the WAL, for Postgres), harmlessly: a restore targets a
// position, never a name.
type DeleteRestorePointReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	Name         string `json:"-"`
}

func NewDeleteRestorePointReq() *DeleteRestorePointReq {
	return &DeleteRestorePointReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *DeleteRestorePointReq) Validate() hperrors.ValidationErrors {
	return validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID)
}

type DeleteRestorePointResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type ManagedRestorePointResp struct {
	Name string `json:"name"`
	// Position is where the point is, as the engine wrote it: a Postgres LSN.
	Position string    `json:"position"`
	At       time.Time `json:"at"`
	// By is the user who made it.
	By        *basedto.ObjectIDResp `json:"by"`
	Note      string                `json:"note"`
	KeepUntil *time.Time            `json:"keepUntil"`
	// Kept says retention is held back for it now.
	Kept bool `json:"kept"`
	// RestorableUntil is until when it can be restored, as retention's next runs
	// go; null when no base backup kept reaches it, or the schedule says nothing.
	RestorableUntil *time.Time `json:"restorableUntil"`
	// Origin is what made it: person, or deploy for a deployment of an app that
	// asks for a point before each of its deploys, which the engine removes once
	// it cannot be restored.
	Origin string `json:"origin"`
}

func TransformRestorePoint(point *entity.ManagedRestorePoint) *ManagedRestorePointResp {
	resp := &ManagedRestorePointResp{Name: point.Name, Position: point.Position, At: point.At, Note: point.Note,
		KeepUntil: timeOrNil(point.KeepUntil), Origin: pointOrigin(point.Origin)}
	if point.By != "" {
		resp.By = &basedto.ObjectIDResp{ID: point.By}
	}
	return resp
}

// TransformRestorePointStatus is a point as Status lists it: with whether it is
// kept now, and until when it can be restored.
func TransformRestorePointStatus(status *managedmodel.RestorePointStatus) *ManagedRestorePointResp {
	resp := TransformRestorePoint(status.Point)
	resp.Kept, resp.RestorableUntil = status.Kept, status.RestorableUntil
	return resp
}

// pointOrigin is a point's origin on the wire: person when none is recorded.
func pointOrigin(origin string) string {
	if origin == "" {
		return pointOriginPerson
	}
	return origin
}

const pointOriginPerson = "person"
