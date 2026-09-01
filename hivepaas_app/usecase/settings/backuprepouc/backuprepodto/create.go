package backuprepodto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/backup"
)

type CreateBackupRepoReq struct {
	settings.CreateSettingReq
	*BackupRepoBaseReq
}

type BackupRepoBaseReq struct {
	Name           string                    `json:"name"`
	Engine         backup.EngineType         `json:"engine"`
	ImportExisting bool                      `json:"importExisting"`
	Description    string                    `json:"description"`
	Password       string                    `json:"password"`
	CloudStorage   basedto.ObjectIDReq       `json:"cloudStorage"`
	Volume         basedto.ObjectIDReq       `json:"volume"`
	StoragePrefix  string                    `json:"storagePrefix"`
	Compression    string                    `json:"compression"`
	PackSize       unit.DataSize             `json:"packSize"`
	Retention      *BackupRetentionPolicyReq `json:"retention"`
}

func (req *BackupRepoBaseReq) ToEntity() *entity.BackupRepo {
	res := &entity.BackupRepo{
		Engine:        req.Engine,
		Description:   req.Description,
		Password:      entity.NewEncryptedField(req.Password),
		CloudStorage:  entity.ObjectID{ID: req.CloudStorage.ID},
		Volume:        entity.ObjectID{ID: req.Volume.ID},
		StoragePrefix: strings.TrimSpace(req.StoragePrefix),
		Compression:   req.Compression,
		PackSize:      req.PackSize,
		Retention:     req.Retention.ToEntity(),
	}
	return res
}

func (req *BackupRepoBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Engine, true, backup.AllEngineTypes, field+"engine")...)
	res = append(res, req.Retention.validate(field+"retention")...)
	return res
}

type BackupRetentionPolicyReq struct {
	KeepLast    int `json:"keepLast"`
	KeepDaily   int `json:"keepDaily"`
	KeepWeekly  int `json:"keepWeekly"`
	KeepMonthly int `json:"keepMonthly"`
}

func (req *BackupRetentionPolicyReq) ToEntity() *entity.BackupRetentionPolicy {
	if req == nil {
		return nil
	}
	return &entity.BackupRetentionPolicy{
		KeepLast:    req.KeepLast,
		KeepDaily:   req.KeepDaily,
		KeepWeekly:  req.KeepWeekly,
		KeepMonthly: req.KeepMonthly,
	}
}

func (req *BackupRetentionPolicyReq) validate(_ string) (res []vld.Validator) {
	return res
}

func NewCreateBackupRepoReq() *CreateBackupRepoReq {
	return &CreateBackupRepoReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *CreateBackupRepoReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateBackupRepoResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
