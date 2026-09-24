package registrydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	domainMaxLen = 255
	idMaxLen     = 50

	maxKeepLast = 1000
	maxKeepDays = 3650

	// maxMemoryLimit is a sanity bound, not a judgement about the right size: it
	// exists so a typo cannot ask swarm for a limit no node could satisfy,
	// leaving the task unschedulable.
	maxMemoryLimit = 64 * unit.GB
	// maxCPULimit is the same kind of bound, in cores.
	maxCPULimit = 256
)

type UpdateRegistrySettingsReq struct {
	settings.UpdateUniqueSettingReq
	*UpdateSettingsBaseReq
}

func NewUpdateRegistrySettingsReq() *UpdateRegistrySettingsReq {
	return &UpdateRegistrySettingsReq{UpdateSettingsBaseReq: &UpdateSettingsBaseReq{}}
}

type UpdateSettingsBaseReq struct {
	Enabled bool       `json:"enabled"`
	Domain  string     `json:"domain"`
	Storage StorageReq `json:"storage"`
	Cleanup CleanupReq `json:"cleanup"`
	// DashboardEnabled serves zot's own web interface at the registry's domain.
	DashboardEnabled bool `json:"dashboardEnabled"`
	// CPULimit and MemoryLimit are what the registry's app runs under. They are
	// not stored with the settings: see Resources. CPULimit is in cores;
	// MemoryLimit is written the way every other size in HivePaaS is - "512mb".
	// Zero in either is the default.
	CPULimit    float64       `json:"cpuLimit"`
	MemoryLimit unit.DataSize `json:"memoryLimit"`

	// RemoveApp and RemoveStorage are asked of this request, not stored by it.
	// Switching the registry off takes its app down, and the app has no screen of
	// its own to be removed from, so the confirmation arrives here.
	RemoveApp bool `json:"removeApp,omitempty"`
	// RemoveStorage deletes the images with the app: the registry's own directory
	// inside the volume, not the volume. A registry on S3 ignores it.
	RemoveStorage bool `json:"removeStorage,omitempty"`
}

type StorageReq struct {
	Type         base.RegistryStorageType `json:"type"`
	Volume       *ObjectIDReq             `json:"volume,omitempty"`
	CloudStorage *ObjectIDReq             `json:"cloudStorage,omitempty"`
}

type CleanupReq struct {
	Enabled  bool `json:"enabled"`
	KeepLast int  `json:"keepLast"`
	KeepDays int  `json:"keepDays"`
}

type ObjectIDReq struct {
	ID string `json:"id"`
}

func (req *ObjectIDReq) toEntity() entity.ObjectID {
	if req == nil {
		return entity.ObjectID{}
	}
	return entity.ObjectID{ID: req.ID}
}

// ToEntity builds what is stored. A nil request is nil rather than a panic: every
// other settings DTO in this repo learned that the hard way, where a missing body
// turned a validation error into a 500.
func (req *UpdateSettingsBaseReq) ToEntity() *entity.RegistrySettings {
	if req == nil {
		return nil
	}
	return &entity.RegistrySettings{
		Enabled: req.Enabled,
		// Type and Managed are not the client's to choose while zot is the only
		// registry HivePaaS provisions. They are written here so that the stored
		// row is complete, and so that a second type later is a wire change
		// rather than a migration.
		Type:    base.RegistryTypeZot,
		Managed: true,
		Domain:  req.Domain,
		Storage: entity.RegistryStorage{
			Type:         req.Storage.Type,
			Volume:       req.Storage.Volume.toEntity(),
			CloudStorage: req.Storage.CloudStorage.toEntity(),
		},
		Cleanup: entity.RegistryCleanup{
			Enabled:  req.Cleanup.Enabled,
			Mode:     base.RegistryCleanupModePolicy,
			KeepLast: req.Cleanup.KeepLast,
			KeepDays: req.Cleanup.KeepDays,
		},
		DashboardEnabled: req.DashboardEnabled,
	}
}

// Resources is what the registry's app is to run under. It goes to the app's
// service, where the app's own resource screen reads and writes it too, rather
// than into the settings. Zero in a field is the default.
func (req *UpdateSettingsBaseReq) Resources() *systemappservice.Resources {
	res := &systemappservice.Resources{
		CPULimit:    req.CPULimit,
		MemoryLimit: req.MemoryLimit.Bytes(),
	}
	if res.CPULimit == 0 {
		res.CPULimit = entity.DefaultRegistryCPULimit
	}
	if res.MemoryLimit == 0 {
		res.MemoryLimit = entity.DefaultRegistryMemoryLimit.Bytes()
	}
	return res
}

func (req *UpdateRegistrySettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	if req.UpdateSettingsBaseReq == nil {
		return hperrors.NewValidationErrors(vld.Validate(validators...))
	}

	validators = append(validators,
		vld.StrLen(&req.Domain, 0, domainMaxLen).OnError(
			vld.SetField("domain", nil),
		),
		vld.NumLTE(&req.Cleanup.KeepLast, maxKeepLast).OnError(
			vld.SetField("cleanup.keepLast", nil),
		),
		vld.NumLTE(&req.Cleanup.KeepDays, maxKeepDays).OnError(
			vld.SetField("cleanup.keepDays", nil),
		),
	)
	if req.Storage.Volume != nil {
		validators = append(validators, vld.StrLen(&req.Storage.Volume.ID, 0, idMaxLen).OnError(
			vld.SetField("storage.volume.id", nil),
		))
	}
	if req.Storage.CloudStorage != nil {
		validators = append(validators, vld.StrLen(&req.Storage.CloudStorage.ID, 0, idMaxLen).OnError(
			vld.SetField("storage.cloudStorage.id", nil),
		))
	}
	// Zero is the default in both; anything else for memory has to leave zot room
	// to run.
	validators = append(validators, basedto.ValidateNumber(&req.MemoryLimit, false,
		entity.MinRegistryMemoryLimit, maxMemoryLimit, "memoryLimit")...)
	validators = append(validators,
		basedto.ValidateNumber(&req.CPULimit, false, 0, maxCPULimit, "cpuLimit")...)

	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRegistrySettingsResp struct {
	Meta *basedto.Meta      `json:"meta"`
	Data *UpdateRegistryRes `json:"data"`
}

// UpdateRegistryRes is what the save did beyond storing the configuration, which
// the screen has no other way of learning.
type UpdateRegistryRes struct {
	// RemovedApp says the registry's app was taken down by this save.
	RemovedApp bool `json:"removedApp"`
	// CredentialKept says the registry account was left in place because an app
	// still names it. It is there so the screen can say where to find it.
	CredentialKept bool `json:"credentialKept"`
}
