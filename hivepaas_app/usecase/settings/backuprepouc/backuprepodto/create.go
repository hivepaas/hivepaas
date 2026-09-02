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

// BackupRepoBaseReq holds what can only be decided when the repository is created. Changing any
// of it later would point at a different repository, or at one in a format it is not in.
type BackupRepoBaseReq struct {
	BackupRepoBaseUpdateReq
	Engine         backup.EngineType   `json:"engine"`
	ImportExisting bool                `json:"importExisting"`
	Password       string              `json:"password"`
	CloudStorage   basedto.ObjectIDReq `json:"cloudStorage"`
	Volume         basedto.ObjectIDReq `json:"volume"`
	StoragePrefix  string              `json:"storagePrefix"`
}

func (req *BackupRepoBaseReq) ToEntity() *entity.BackupRepo {
	res := &entity.BackupRepo{
		Engine:        req.Engine,
		Password:      entity.NewEncryptedField(req.Password),
		CloudStorage:  entity.ObjectID{ID: req.CloudStorage.ID},
		Volume:        entity.ObjectID{ID: req.Volume.ID},
		StoragePrefix: strings.TrimSpace(req.StoragePrefix),
	}
	req.Apply(res)
	return res
}

func (req *BackupRepoBaseReq) validate(field string) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, req.BackupRepoBaseUpdateReq.validate(field, req.Engine)...)
	res = append(res, basedto.ValidateStrIn(&req.Engine, true, backup.AllEngineTypes, field+"engine")...)
	return res
}

// BackupRepoBaseUpdateReq holds what stays changeable once the repository exists.
type BackupRepoBaseUpdateReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Compression and PackSize are stored inside the repository itself, so changing them here
	// only takes effect once they are pushed to it. See UpdateBackupRepo.
	Compression string                    `json:"compression"`
	PackSize    unit.DataSize             `json:"packSize"`
	Retention   *BackupRetentionPolicyReq `json:"retention"`
}

func (req *BackupRepoBaseUpdateReq) Apply(repo *entity.BackupRepo) {
	repo.Description = req.Description
	repo.Compression = req.Compression
	repo.Retention = req.Retention.ToEntity()

	// An unset pack size means "leave the repository alone" - the engine has no way to unset it,
	// so storing zero here would leave the setting disagreeing with the repository for good.
	if req.PackSize > 0 {
		repo.PackSize = req.PackSize
	}
}

// RepoOptions returns the subset that has to be pushed to the repository to take effect,
// resolved against what the repository currently has so an omitted pack size does not read as
// a change that can never be applied.
func (req *BackupRepoBaseUpdateReq) RepoOptions(repo *entity.BackupRepo) backup.RepoOptions {
	packSize := req.PackSize
	if packSize <= 0 {
		packSize = repo.PackSize
	}
	return backup.NewRepoOptions(int(packSize.MBytes()), req.Compression)
}

func (req *BackupRepoBaseUpdateReq) validate(field string, engine backup.EngineType) (res []vld.Validator) {
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.Name, true, 1, base.SettingNameMaxLen, field+"name")...)
	res = append(res, basedto.ValidateStrIn(&req.Compression, false,
		backup.AllCompressionAlgorithms[engine], field+"compression")...)
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
	validators = append(validators, req.CreateSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type CreateBackupRepoResp struct {
	Meta *basedto.Meta         `json:"meta"`
	Data *basedto.ObjectIDResp `json:"data"`
}
