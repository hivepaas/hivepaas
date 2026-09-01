package backuprepodto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/backup"
)

const (
	maskedSecret = "********"
)

type GetBackupRepoReq struct {
	settings.GetSettingReq
}

func NewGetBackupRepoReq() *GetBackupRepoReq {
	return &GetBackupRepoReq{}
}

func (req *GetBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.GetSettingReq.Validate()...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetBackupRepoResp struct {
	Meta *basedto.Meta   `json:"meta"`
	Data *BackupRepoResp `json:"data"`
}

type BackupRepoResp struct {
	*settings.BaseSettingResp
	Engine        backup.EngineType         `json:"engine"`
	Description   string                    `json:"description,omitempty"`
	Password      string                    `json:"password,omitempty"`
	CloudStorage  *settings.BaseSettingResp `json:"cloudStorage,omitempty" copy:"-"`
	Volume        *settings.BaseSettingResp `json:"volume,omitempty" copy:"-"`
	StoragePrefix string                    `json:"storagePrefix,omitempty"`
	Compression   string                    `json:"compression,omitempty"`
	PackSize      unit.DataSize             `json:"packSize,omitempty"`
	Retention     *BackupRetentionPolicyReq `json:"retention,omitempty"`
	SecretMasked  bool                      `json:"secretMasked,omitempty"`
}

func (resp *BackupRepoResp) CopyPassword(field entity.EncryptedField) error {
	resp.Password = field.String()
	return nil
}

type BackupRetentionPolicyResp struct {
	KeepLast    int `json:"keepLast,omitempty"`
	KeepDaily   int `json:"keepDaily,omitempty"`
	KeepWeekly  int `json:"keepWeekly,omitempty"`
	KeepMonthly int `json:"keepMonthly,omitempty"`
}

func TransformBackupRepo(
	setting *entity.Setting,
	refObjects *entity.RefObjects,
) (resp *BackupRepoResp, err error) {
	config := setting.MustAsBackupRepo()
	if err = copier.Copy(&resp, &config); err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if config.CloudStorage.ID != "" {
		sourceSetting := refObjects.RefSettings[config.CloudStorage.ID]
		storageResp, _ := settings.TransformSettingBase(sourceSetting)
		if storageResp == nil {
			storageResp = settings.NewMissingSetting(config.CloudStorage.ID, base.SettingTypeCloudStorage)
		}
		resp.CloudStorage = storageResp
	}
	if config.Volume.ID != "" {
		sourceSetting := refObjects.RefSettings[config.Volume.ID]
		storageResp, _ := settings.TransformSettingBase(sourceSetting)
		if storageResp == nil {
			storageResp = settings.NewMissingSetting(config.Volume.ID, base.SettingTypeClusterVolume)
		}
		resp.Volume = storageResp
	}

	resp.SecretMasked = resp.Inherited
	if resp.SecretMasked {
		resp.Password = maskedSecret
	}

	return resp, nil
}
