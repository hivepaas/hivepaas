package appmanageddto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	maxRetainFull          = 365
	restoreTestQueryMaxLen = 2000
)

type GetManagedSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetManagedSettingsReq() *GetManagedSettingsReq {
	return &GetManagedSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetManagedSettingsReq) Validate() hperrors.ValidationErrors {
	return validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID)
}

func validateAppPath(projectID, projectEnvID, appID string) hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&projectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&projectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&appID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetManagedSettingsResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *ManagedSettingsResp `json:"data"`
}

// ManagedSettingsResp is what a managed app's engine keeps of it: where it
// archives, how long it keeps what, when its jobs run, and what it owns of the
// app, which the settings screens show read-only.
type ManagedSettingsResp struct {
	UpdateVer           int                       `json:"updateVer"`
	Engine              string                    `json:"engine"`
	EngineAPI           int                       `json:"engineApi"`
	Storage             *ManagedStorageResp       `json:"storage"`
	ArchiveTimeout      timeutil.Duration         `json:"archiveTimeout"`
	RetainFull          int                       `json:"retainFull"`
	BaseBackupSchedule  string                    `json:"baseBackupSchedule"`
	RestoreTestSchedule string                    `json:"restoreTestSchedule"`
	RestoreTest         *ManagedRestoreTestCfgRsp `json:"restoreTest"`
	Locks               *ManagedLocksResp         `json:"locks"`
	Jobs                *ManagedJobsResp          `json:"jobs"`
}

type ManagedStorageResp struct {
	BackupRepo   *basedto.ObjectIDResp `json:"backupRepo"`
	CloudStorage *basedto.ObjectIDResp `json:"cloudStorage"`
	KeyAuth      *basedto.ObjectIDResp `json:"keyAuth"`
	Prefix       string                `json:"prefix"`
	Placement    string                `json:"placement"`
}

type ManagedRestoreTestCfgRsp struct {
	// Mode is container or verifier.
	Mode        string                `json:"mode"`
	VerifierApp *basedto.ObjectIDResp `json:"verifierApp"`
	NodeID      string                `json:"nodeId"`
	CheckQuery  string                `json:"checkQuery"`
}

type ManagedLocksResp struct {
	Image     string   `json:"image"`
	ImageLine string   `json:"imageLine"`
	Command   string   `json:"command"`
	Mounts    []string `json:"mounts"`
	EnvKeys   []string `json:"envKeys"`
	Settings  []string `json:"settings"`
}

type ManagedJobsResp struct {
	Nightly     *basedto.ObjectIDResp `json:"nightly"`
	Health      *basedto.ObjectIDResp `json:"health"`
	RestoreTest *basedto.ObjectIDResp `json:"restoreTest"`
}

// ManagedSettingsView is what the settings are read from: the app-managed
// setting, parsed, and the jobs its engine made, by id.
type ManagedSettingsView struct {
	Setting *entity.Setting
	Managed *entity.AppManagedSettings
	Jobs    map[string]*entity.SchedJob
}

func TransformManagedSettings(view *ManagedSettingsView) *ManagedSettingsResp {
	managed := view.Managed
	locks := managed.Locks
	resp := &ManagedSettingsResp{
		UpdateVer: view.Setting.UpdateVer,
		Engine:    managed.Engine,
		EngineAPI: managed.EngineAPI,
		Locks: &ManagedLocksResp{Image: locks.Image, ImageLine: locks.ImageLine, Command: locks.Command,
			Mounts: nonNil(locks.Mounts), EnvKeys: nonNil(locks.EnvKeys), Settings: nonNil(locks.Settings)},
	}
	pg := managed.Postgres
	if pg == nil {
		return resp
	}
	resp.Storage = &ManagedStorageResp{
		BackupRepo:   objectID(pg.Storage.BackupRepo),
		CloudStorage: objectID(pg.Storage.CloudStorage),
		KeyAuth:      objectID(pg.Storage.KeyAuth),
		Prefix:       pg.Storage.Prefix,
		Placement:    pg.Storage.Placement,
	}
	resp.ArchiveTimeout = pg.ArchiveTimeout
	resp.RetainFull = pg.RetainFull
	resp.RestoreTest = &ManagedRestoreTestCfgRsp{Mode: pg.RestoreTest.Mode,
		VerifierApp: objectID(pg.RestoreTest.VerifierApp), NodeID: pg.RestoreTest.NodeID,
		CheckQuery: pg.RestoreTest.CheckQuery}
	resp.Jobs = &ManagedJobsResp{Nightly: objectID(pg.Jobs.Nightly), Health: objectID(pg.Jobs.Health),
		RestoreTest: objectID(pg.Jobs.RestoreTest)}
	resp.BaseBackupSchedule = cronOf(view.Jobs[pg.Jobs.Nightly.ID])
	resp.RestoreTestSchedule = cronOf(view.Jobs[pg.Jobs.RestoreTest.ID])
	return resp
}

func objectID(id entity.ObjectID) *basedto.ObjectIDResp {
	if id.ID == "" {
		return nil
	}
	return &basedto.ObjectIDResp{ID: id.ID}
}

func cronOf(job *entity.SchedJob) string {
	if job == nil || job.Schedule == nil {
		return ""
	}
	return job.Schedule.CronExpr
}

func nonNil(items []string) []string {
	if items == nil {
		return []string{}
	}
	return items
}

// UpdateManagedSettingsReq changes a managed app's settings: a field left out is
// left as it is, and an empty query clears it.
type UpdateManagedSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	UpdateVer           int     `json:"updateVer"`
	RetainFull          *int    `json:"retainFull,omitempty"`
	BaseBackupSchedule  *string `json:"baseBackupSchedule,omitempty"`
	RestoreTestSchedule *string `json:"restoreTestSchedule,omitempty"`
	RestoreTestQuery    *string `json:"restoreTestQuery,omitempty"`
}

func NewUpdateManagedSettingsReq() *UpdateManagedSettingsReq {
	return &UpdateManagedSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateManagedSettingsReq) Validate() hperrors.ValidationErrors {
	if errs := validateAppPath(req.ProjectID, req.ProjectEnvID, req.AppID); len(errs) > 0 {
		return errs
	}
	validators := make([]vld.Validator, 0, 2) //nolint:mnd
	if req.RetainFull != nil {
		validators = append(validators,
			basedto.ValidateCond(*req.RetainFull >= 1 && *req.RetainFull <= maxRetainFull, "retainFull")...)
	}
	if req.RestoreTestQuery != nil {
		validators = append(validators,
			basedto.ValidateCond(len(*req.RestoreTestQuery) <= restoreTestQueryMaxLen, "restoreTestQuery")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateManagedSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
