package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentAppKindSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppKind, &appKindSettingsParser{})

type appKindSettingsParser struct {
}

func (s *appKindSettingsParser) New() SettingData {
	return &AppKindSettings{}
}

type AppKindSettings struct {
	Category base.AppCategory `json:"category"`          // database, webapp, cache...
	Engine   string           `json:"engine"`            // postgres, mysql, redis, nodejs, python...
	Version  string           `json:"version,omitempty"` // 16, 8.0, 1.2-alpine ...

	Database *AppKindDatabase `json:"database,omitempty"`
	Webapp   *AppKindWebapp   `json:"webapp,omitempty"`
	Cache    *AppKindCache    `json:"cache,omitempty"`
	Storage  *AppKindStorage  `json:"storage,omitempty"`
}

type AppKindDatabase struct {
	DbName       string               `json:"dbName,omitempty"`
	Username     string               `json:"username,omitempty"`
	Password     EncryptedField       `json:"password,omitzero"`
	RootPassword EncryptedField       `json:"rootPassword,omitzero"`
	SSLMode      base.DatabaseSSLMode `json:"sslMode,omitempty"` // disable | require | verify-full ...
}

type AppKindWebapp struct {
}

type AppKindCache struct {
	Password        EncryptedField `json:"password,omitzero"`
	MaxMemory       unit.DataSize  `json:"maxMemory,omitempty"`
	EvictionRule    string         `json:"evictionRule,omitempty"`
	PersistenceMode string         `json:"persistenceMode,omitempty"`
}

// AppKindStorage is a store other apps put objects in: the credentials they
// authenticate with, and where what they send lands.
//
// KeyID and Secret are named for what they are rather than for S3, because not
// every store speaks S3: they carry an access key and a secret key for
// SeaweedFS or MinIO, and an ordinary user and password for a server that
// authenticates over HTTP, such as restic's.
type AppKindStorage struct {
	KeyID  string         `json:"keyId,omitempty"`
	Secret EncryptedField `json:"secret,omitzero"`
	Bucket string         `json:"bucket,omitempty"`
	Region string         `json:"region,omitempty"`
}

func (s *AppKindSettings) GetType() base.SettingType {
	return base.SettingTypeAppKind
}

func (s *AppKindSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppKindSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AppKindSettings) Decrypt() error {
	if s.Database != nil {
		if _, err := s.Database.Password.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
		if _, err := s.Database.RootPassword.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if s.Cache != nil {
		if _, err := s.Cache.Password.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	if s.Storage != nil {
		if _, err := s.Storage.Secret.GetPlain(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (s *Setting) AsAppKindSettings() (*AppKindSettings, error) {
	return parseSettingAs[*AppKindSettings](s)
}

func (s *Setting) MustAsAppKindSettings() *AppKindSettings {
	return gofn.Must(s.AsAppKindSettings())
}
