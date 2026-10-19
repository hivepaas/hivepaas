package apptemplatedto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	appNameMaxLen      = 100
	choiceNameMaxLen   = 32
	templateNameMinLen = 1
	// imageTagMaxLen is docker's own limit on a tag.
	imageTagMaxLen = 128
	// maxRetainFull and restoreTestQueryMaxLen bound a managed app's options.
	maxRetainFull          = 365
	restoreTestQueryMaxLen = 2000
	cronMaxLen             = 100
)

type CreateAppFromTemplateReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`

	Name     string `json:"name"`
	Template string `json:"template"`
	// Version and Variant are empty for the template's defaults.
	Version string `json:"version"`
	Variant string `json:"variant"`
	// ImageTag replaces the tag of the image the template's version pins, for this
	// app only - one of the tags the image-tags endpoint lists. The repository is
	// the template's and is not asked for: everything else the template says is
	// only correct for that software.
	ImageTag string `json:"imageTag"`
	// Params are JSON values of each parameter's type; a size is a string such as "1GB".
	Params map[string]any `json:"params"`
	// DependencyParams are the values a person gave for each dependency the
	// template declares, by the dependency's name: in practice its data volume.
	DependencyParams map[string]map[string]any `json:"dependencyParams"`
	// ResetStorage deletes what a previous install of these apps left in their
	// directories before the new ones are created. It is what the preflight
	// endpoint's findings are answered with, and it is off unless asked for.
	ResetStorage bool `json:"resetStorage,omitempty"`
	// Managed is how the engine of a managed template looks after the app: where
	// it archives, and what the template's defaults are changed to. Required for
	// a managed template, and refused for any other.
	Managed *ManagedCreateReq `json:"managed,omitempty"`
}

// ManagedCreateReq is what a managed app is created with. A zero field takes
// the template's default, then the engine's.
type ManagedCreateReq struct {
	Storage             ManagedStorageReq `json:"storage"`
	ArchiveTimeout      timeutil.Duration `json:"archiveTimeout,omitempty"`
	RetainFull          int               `json:"retainFull,omitempty"`
	BaseBackupSchedule  string            `json:"baseBackupSchedule,omitempty"`
	RestoreTestSchedule string            `json:"restoreTestSchedule,omitempty"`
	RestoreTestQuery    string            `json:"restoreTestQuery,omitempty"`
	// Dump is the database's logical dump (pg_dump, mariadb-dump) into a backup
	// repository. Left out, a
	// database that archives into a backup repository is dumped into it too, and
	// one that archives into cloud storage is not dumped until a repository is
	// chosen.
	Dump *ManagedDumpReq `json:"dump,omitempty"`
}

// ManagedDumpReq is where and when the database is dumped: a repository of null
// makes no dumps, and an empty schedule is the engine's default.
type ManagedDumpReq struct {
	Repo     *basedto.ObjectIDReq `json:"repo"`
	Schedule string               `json:"schedule,omitempty"`
}

// ManagedStorageReq is where the engine archives: a backup repository, S3 or
// volume, or any cloud storage - exactly one. KeyAuth is a narrower key than the
// storage's own, for an S3 archive.
type ManagedStorageReq struct {
	BackupRepo   *basedto.ObjectIDReq `json:"backupRepo,omitempty"`
	CloudStorage *basedto.ObjectIDReq `json:"cloudStorage,omitempty"`
	KeyAuth      *basedto.ObjectIDReq `json:"keyAuth,omitempty"`
}

func NewCreateAppFromTemplateReq() *CreateAppFromTemplateReq {
	return &CreateAppFromTemplateReq{}
}

// ModifyRequest implements interface basedto.ReqModifier
func (req *CreateAppFromTemplateReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Template = strings.TrimSpace(req.Template)
	req.Version = strings.TrimSpace(req.Version)
	req.Variant = strings.TrimSpace(req.Variant)
	req.ImageTag = strings.TrimSpace(req.ImageTag)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CreateAppFromTemplateReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 8) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, appNameMaxLen, "name")...)
	validators = append(validators,
		basedto.ValidateStr(&req.Template, true, templateNameMinLen, templateNameMaxLen, "template")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	validators = append(validators, basedto.ValidateStr(&req.ImageTag, false, 1, imageTagMaxLen, "imageTag")...)
	validators = append(validators,
		basedto.ValidateCond(req.ImageTag == "" || templatemodel.ValidImageTag(req.ImageTag), "imageTag")...)
	validators = append(validators,
		basedto.ValidateCond(len(req.DependencyParams) <= templatemodel.MaxDependencies, "dependencyParams")...)
	if managed := req.Managed; managed != nil {
		storages := 0
		for _, ref := range []*basedto.ObjectIDReq{managed.Storage.BackupRepo, managed.Storage.CloudStorage} {
			if ref != nil && ref.ID != "" {
				storages++
			}
		}
		validators = append(validators,
			basedto.ValidateCond(storages == 1, "managed.storage")...)
		validators = append(validators,
			basedto.ValidateCond(managed.RetainFull >= 0 && managed.RetainFull <= maxRetainFull, "managed.retainFull")...)
		validators = append(validators, basedto.ValidateStr(&managed.RestoreTestQuery, false, 1,
			restoreTestQueryMaxLen, "managed.restoreTestQuery")...)
		if dump := managed.Dump; dump != nil {
			if dump.Repo != nil && dump.Repo.ID != "" {
				validators = append(validators, basedto.ValidateObjectIDReq(dump.Repo, true, "managed.dump.repo")...)
			}
			validators = append(validators,
				basedto.ValidateStr(&dump.Schedule, false, 1, cronMaxLen, "managed.dump.schedule")...)
		}
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateAppFromTemplateResp struct {
	Meta *basedto.Meta                  `json:"meta"`
	Data *CreateAppFromTemplateDataResp `json:"data"`
}

type CreateAppFromTemplateDataResp struct {
	App *basedto.ObjectIDResp `json:"app"`
	// Deployment is the app's first deployment, already queued: it pulls the
	// template's image.
	Deployment *basedto.ObjectIDResp `json:"deployment"`
	// Dependencies are the apps created alongside, each with its first deployment.
	Dependencies []*CreatedDependencyResp `json:"dependencies"`
	// Components are the other apps of an application that is several processes,
	// each with its first deployment. The app above is the primary component, so
	// it is not repeated here.
	Components []*CreatedComponentResp `json:"components"`
	// Managed is what the engine of a managed template set up, and null for any
	// other template.
	Managed *CreatedManagedResp `json:"managed"`
}

// CreatedManagedResp says where the archive is relative to the database:
// off-site, on storage the database's node shares, or on the same node, which
// the dashboard warns about - losing the node loses both.
type CreatedManagedResp struct {
	Placement string `json:"placement"`
}

type CreatedComponentResp struct {
	// Name is the component's role in the template: auth, worker.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}

type CreatedDependencyResp struct {
	// Name is the dependency's role in the template: db, cache.
	Name       string                `json:"name"`
	App        *basedto.ObjectIDResp `json:"app"`
	Deployment *basedto.ObjectIDResp `json:"deployment"`
}
