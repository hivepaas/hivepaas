package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	nameMaxLen     = 100
	versionMaxLen  = 50
	cacheConfigMax = 100
	portMax        = 65535
)

type UpdateAppKindSettingsReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
	*AppKindSettingsReq
}

type AppKindSettingsReq struct {
	Category base.AppCategory `json:"category"`
	Engine   string           `json:"engine"`
	Port     uint             `json:"port"`
	Version  string           `json:"version"`

	Webapp   *AppKindWebappReq   `json:"webapp"`
	Database *AppKindDatabaseReq `json:"database"`
	Cache    *AppKindCacheReq    `json:"cache"`
	Storage  *AppKindStorageReq  `json:"storage"`

	UpdateVer int `json:"updateVer"`
}

// KeepMaskedSecrets restores the stored value for the secret the request only
// carries as the masked placeholder the GET response substitutes for it.
func (req *AppKindSettingsReq) KeepMaskedSecrets(newSettings, current *entity.AppKindSettings) {
	if newSettings == nil || current == nil {
		return
	}
	if newSettings.Database != nil && current.Database != nil && req.Database != nil {
		if basedto.IsMaskedSecret(req.Database.Password) {
			newSettings.Database.Password = current.Database.Password
		}
		if basedto.IsMaskedSecret(req.Database.RootPassword) {
			newSettings.Database.RootPassword = current.Database.RootPassword
		}
	}
	if newSettings.Cache != nil && current.Cache != nil && req.Cache != nil {
		if basedto.IsMaskedSecret(req.Cache.Password) {
			newSettings.Cache.Password = current.Cache.Password
		}
	}
	if newSettings.Storage != nil && current.Storage != nil && req.Storage != nil {
		if basedto.IsMaskedSecret(req.Storage.Secret) {
			newSettings.Storage.Secret = current.Storage.Secret
		}
	}
}

func (req *AppKindSettingsReq) ToEntity() *entity.AppKindSettings {
	return &entity.AppKindSettings{
		Category: req.Category,
		Engine:   req.Engine,
		Version:  req.Version,

		Database: req.Database.ToEntity(),
		Webapp:   req.Webapp.ToEntity(),
		Cache:    req.Cache.ToEntity(),
		Storage:  req.Storage.ToEntity(),
	}
}

func (req *AppKindSettingsReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStrIn(&req.Category, true, base.AllAppCategories, field+"category")...)
	res = append(res, basedto.ValidateStr(&req.Engine, false, 1, nameMaxLen, field+"engine")...)
	res = append(res, basedto.ValidateNumber(&req.Port, false, 1, portMax, field+"port")...)
	res = append(res, basedto.ValidateStr(&req.Version, false, 1, versionMaxLen, field+"version")...)
	switch req.Category {
	case base.AppCategoryWebapp:
		res = append(res, basedto.ValidateCond(req.Webapp != nil, field+"webapp")...)
		res = append(res, req.Webapp.validate(field+"webapp")...)
	case base.AppCategoryDatabase:
		res = append(res, basedto.ValidateCond(req.Database != nil, field+"database")...)
		res = append(res, req.Database.validate(field+"database")...)
	case base.AppCategoryCache:
		res = append(res, basedto.ValidateCond(req.Cache != nil, field+"cache")...)
		res = append(res, req.Cache.validate(field+"cache")...)
	case base.AppCategoryStorage:
		res = append(res, basedto.ValidateCond(req.Storage != nil, field+"storage")...)
		res = append(res, req.Storage.validate(field+"storage")...)
	}
	return res
}

type AppKindWebappReq struct {
}

func (req *AppKindWebappReq) ToEntity() *entity.AppKindWebapp {
	if req == nil {
		return nil
	}
	return &entity.AppKindWebapp{}
}

//nolint:unparam
func (req *AppKindWebappReq) validate(_ string) (res []vld.Validator) {
	if req == nil {
		return
	}
	// TODO: add validation
	return res
}

type AppKindDatabaseReq struct {
	DbName         string               `json:"dbName"`
	Username       string               `json:"username"`
	Password       string               `json:"password"`
	RootPassword   string               `json:"rootPassword"`
	SSLMode        base.DatabaseSSLMode `json:"sslMode"`
	SSLCert        basedto.ObjectIDReq  `json:"sslCert"`
	TLSPassthrough bool                 `json:"tlsPassthrough"`
}

func (req *AppKindDatabaseReq) ToEntity() *entity.AppKindDatabase {
	if req == nil {
		return nil
	}
	return &entity.AppKindDatabase{
		DbName:       req.DbName,
		Username:     req.Username,
		Password:     entity.NewEncryptedField(req.Password),
		RootPassword: entity.NewEncryptedField(req.RootPassword),
		SSLMode:      req.SSLMode,
	}
}

func (req *AppKindDatabaseReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.DbName, false, 1, nameMaxLen, field+"dbName")...)
	res = append(res, basedto.ValidateStr(&req.Username, false, 1, nameMaxLen, field+"username")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	res = append(res, basedto.ValidatePlainSecret(&req.RootPassword, field+"rootPassword")...)
	res = append(res, basedto.ValidateStrIn(&req.SSLMode, true, base.AllDatabaseSslModes, field+"sslMode")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.SSLCert, false, field+"sslCert")...)
	return res
}

type AppKindCacheReq struct {
	Password        string              `json:"password" copy:"-"`
	MaxMemory       unit.DataSize       `json:"maxMemory" swaggertype:"string"`
	EvictionRule    string              `json:"evictionRule"`
	PersistenceMode string              `json:"persistenceMode"`
	SSLCert         basedto.ObjectIDReq `json:"sslCert"`
}

func (req *AppKindCacheReq) ToEntity() *entity.AppKindCache {
	if req == nil {
		return nil
	}
	return &entity.AppKindCache{
		Password:        entity.NewEncryptedField(req.Password),
		MaxMemory:       req.MaxMemory,
		EvictionRule:    req.EvictionRule,
		PersistenceMode: req.PersistenceMode,
	}
}

func (req *AppKindCacheReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidatePlainSecret(&req.Password, field+"password")...)
	res = append(res, basedto.ValidateStr(&req.EvictionRule, false, 1, cacheConfigMax, field+"evictionRule")...)
	res = append(res, basedto.ValidateStr(&req.PersistenceMode, false, 1, cacheConfigMax, field+"persistenceMode")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.SSLCert, false, field+"sslCert")...)
	return res
}

func NewUpdateAppKindSettingsReq() *UpdateAppKindSettingsReq {
	return &UpdateAppKindSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppKindSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppKindSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}

type AppKindStorageReq struct {
	KeyID  string `json:"keyId"`
	Secret string `json:"secret" copy:"-"`
	Bucket string `json:"bucket"`
	Region string `json:"region"`
}

func (req *AppKindStorageReq) ToEntity() *entity.AppKindStorage {
	if req == nil {
		return nil
	}
	return &entity.AppKindStorage{
		KeyID:  req.KeyID,
		Secret: entity.NewEncryptedField(req.Secret),
		Bucket: req.Bucket,
		Region: req.Region,
	}
}

func (req *AppKindStorageReq) validate(field string) (res []vld.Validator) {
	if req == nil {
		return
	}
	if field != "" {
		field += "."
	}
	res = append(res, basedto.ValidateStr(&req.KeyID, false, 1, nameMaxLen, field+"keyId")...)
	res = append(res, basedto.ValidatePlainSecret(&req.Secret, field+"secret")...)
	res = append(res, basedto.ValidateStr(&req.Bucket, false, 1, nameMaxLen, field+"bucket")...)
	res = append(res, basedto.ValidateStr(&req.Region, false, 1, nameMaxLen, field+"region")...)
	return res
}
