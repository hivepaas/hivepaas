package registrydto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
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
	// MemoryLimit is written the way every other size in HivePaaS is - "512mb" -
	// and decoded into the same type the setting stores.
	MemoryLimit unit.DataSize `json:"memoryLimit"`
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
		MemoryLimit:      req.MemoryLimit,
	}
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
	validators = append(validators,
		basedto.ValidateNumber(&req.MemoryLimit, false, 0, maxMemoryLimit, "memoryLimit")...)

	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateRegistrySettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
