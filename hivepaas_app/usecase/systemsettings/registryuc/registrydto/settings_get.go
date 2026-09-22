package registrydto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetRegistrySettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetRegistrySettingsReq() *GetRegistrySettingsReq {
	return &GetRegistrySettingsReq{}
}

func (req *GetRegistrySettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetRegistrySettingsResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *RegistrySettingsResp `json:"data"`
}

type RegistrySettingsResp struct {
	*settings.BaseSettingResp

	Enabled bool              `json:"enabled"`
	Type    base.RegistryType `json:"type"`
	Managed bool              `json:"managed"`
	Domain  string            `json:"domain"`
	Storage StorageResp       `json:"storage"`
	Cleanup CleanupResp       `json:"cleanup"`
	// DashboardEnabled says whether zot's own web interface answers at the
	// registry's domain.
	DashboardEnabled bool                  `json:"dashboardEnabled"`
	MemoryLimit      unit.DataSize         `json:"memoryLimit"`
	App              *basedto.ObjectIDResp `json:"app,omitempty"`
	Credential       *basedto.ObjectIDResp `json:"credential,omitempty"`
	// RegistryStatus is not called Status: BaseSettingResp already has one, and
	// the copier writes the setting's status into whatever field that name
	// matches - which turned the whole response into a 500 the first time it ran.
	RegistryStatus *RegistryStatusResp    `json:"registryStatus"`
	Credentials    *CredentialRotationRes `json:"credentialRotation,omitempty"`
}

type StorageResp struct {
	Type         base.RegistryStorageType `json:"type"`
	Volume       *basedto.ObjectIDResp    `json:"volume,omitempty"`
	CloudStorage *basedto.ObjectIDResp    `json:"cloudStorage,omitempty"`
}

type CleanupResp struct {
	Enabled  bool                     `json:"enabled"`
	Mode     base.RegistryCleanupMode `json:"mode"`
	KeepLast int                      `json:"keepLast"`
	KeepDays int                      `json:"keepDays"`
}

// RegistryStatusResp is what is running, as opposed to what was configured.
type RegistryStatusResp struct {
	Provisioned  bool   `json:"provisioned"`
	AppID        string `json:"appId,omitempty"`
	Reachable    bool   `json:"reachable"`
	Unreachable  string `json:"unreachable,omitempty"`
	Repositories int    `json:"repositories"`
	StoredBytes  int64  `json:"storedBytes"`
}

// CredentialRotationRes says when the password last changed and how long the one
// before it still works, which is what tells an operator whether their apps have
// to be redeployed yet.
type CredentialRotationRes struct {
	RotatedAt   time.Time `json:"rotatedAt,omitzero"`
	GraceEndsAt time.Time `json:"graceEndsAt,omitzero"`
}

type RegistrySettingsTransformationInput struct {
	RegistrySetting *entity.Setting
	RegistryStatus  *registryservice.Status
}

// TransformRegistrySettings writes the stored configuration and what is running
// into one response. The password is not in it: the dashboard links to the
// registry auth setting, which has its own reveal flow.
func TransformRegistrySettings(
	input *RegistrySettingsTransformationInput,
) (*RegistrySettingsResp, error) {
	resp := &RegistrySettingsResp{}
	if input.RegistrySetting == nil {
		return resp, nil
	}

	if err := copier.Copy(&resp, input.RegistrySetting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := input.RegistrySetting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.Enabled, resp.Type, resp.Managed = cfg.Enabled, cfg.Type, cfg.Managed
	resp.Domain, resp.MemoryLimit = cfg.Domain, cfg.MemoryLimit
	resp.DashboardEnabled = cfg.DashboardEnabled
	resp.Storage = StorageResp{Type: cfg.Storage.Type}
	if cfg.Storage.Volume.ID != "" {
		resp.Storage.Volume = &basedto.ObjectIDResp{ID: cfg.Storage.Volume.ID}
	}
	if cfg.Storage.CloudStorage.ID != "" {
		resp.Storage.CloudStorage = &basedto.ObjectIDResp{ID: cfg.Storage.CloudStorage.ID}
	}
	resp.Cleanup = CleanupResp{
		Enabled:  cfg.Cleanup.Enabled,
		Mode:     cfg.Cleanup.Mode,
		KeepLast: cfg.Cleanup.KeepLast,
		KeepDays: cfg.Cleanup.KeepDays,
	}
	if cfg.AppID != "" {
		resp.App = &basedto.ObjectIDResp{ID: cfg.AppID}
	}
	if cfg.RegistryAuthID != "" {
		resp.Credential = &basedto.ObjectIDResp{ID: cfg.RegistryAuthID}
	}
	if !cfg.CredentialRotatedAt.IsZero() {
		resp.Credentials = &CredentialRotationRes{
			RotatedAt:   cfg.CredentialRotatedAt,
			GraceEndsAt: cfg.CredentialGraceEnds,
		}
	}

	resp.RegistryStatus = &RegistryStatusResp{}
	if input.RegistryStatus != nil {
		resp.RegistryStatus.Provisioned = input.RegistryStatus.Provisioned
		resp.RegistryStatus.AppID = input.RegistryStatus.AppID
		resp.RegistryStatus.Reachable = input.RegistryStatus.Reachable
		resp.RegistryStatus.Unreachable = input.RegistryStatus.Unreachable
		resp.RegistryStatus.Repositories = input.RegistryStatus.Repositories
		resp.RegistryStatus.StoredBytes = input.RegistryStatus.StoredBytes
	}
	return resp, nil
}
