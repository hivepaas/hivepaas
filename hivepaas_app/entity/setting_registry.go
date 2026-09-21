package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	CurrentRegistrySettingsVersion = 1

	// DefaultRegistryKeepLast and DefaultRegistryKeepDays are the cleanup an
	// operator gets without saying anything: the last ten builds of every app,
	// and everything from the past month.
	DefaultRegistryKeepLast = 10
	DefaultRegistryKeepDays = 30

	DefaultRegistryMemoryLimit = 512 * unit.MB
	MinRegistryMemoryLimit     = 256 * unit.MB
)

var _ = registerSettingParser(base.SettingTypeRegistry, &registrySettingsParser{})

type registrySettingsParser struct {
}

// New is both the never-configured default and the struct stored data is
// unmarshalled into, so every field defaulted here to a non-zero value must be
// written unconditionally - with `omitempty`, a stored false is left out of the
// JSON and the default below survives the unmarshal as a silent true.
func (s *registrySettingsParser) New() SettingData {
	return &RegistrySettings{
		Type:    base.RegistryTypeZot,
		Managed: true,
		Storage: RegistryStorage{Type: base.RegistryStorageTypeVolume},
		Cleanup: RegistryCleanup{
			Enabled:  true,
			Mode:     base.RegistryCleanupModePolicy,
			KeepLast: DefaultRegistryKeepLast,
			KeepDays: DefaultRegistryKeepDays,
		},
		MemoryLimit: DefaultRegistryMemoryLimit,
	}
}

// RegistrySettings is the whole subsystem's configuration. It is global, and until
// Enabled is set nothing is provisioned.
type RegistrySettings struct {
	// Enabled is what provisions the registry. Clearing it removes nothing: the
	// images are on a volume or in a bucket, and deleting them is done in the
	// app's own screen.
	Enabled bool `json:"enabled,omitempty"`

	// Type and Managed are separate on purpose, the way logging's are: an
	// operator pointing HivePaaS at a registry it does not run is a case worth
	// keeping expressible, even though Managed is the only value today.
	Type base.RegistryType `json:"type,omitempty"`
	// Managed defaults to true in New, so it carries no omitempty.
	Managed bool `json:"managed"`

	// Domain is the address images are named with, and it is required to enable
	// the registry: a docker daemon speaks to a registry over HTTPS at a name.
	Domain string `json:"domain,omitempty"`

	Storage RegistryStorage `json:"storage"`
	Cleanup RegistryCleanup `json:"cleanup"`

	MemoryLimit unit.DataSize `json:"memoryLimit,omitempty"`

	// AppID and RegistryAuthID are what provisioning created. They are how Apply
	// finds its own work again, and how the dashboard links to it.
	AppID          string `json:"appId,omitempty"`
	RegistryAuthID string `json:"registryAuthId,omitempty"`

	// CredentialRotatedAt is when the password last changed, and
	// CredentialGraceEnds is when the one before it stops working. A swarm
	// service presents the credential it was deployed with, so an app nobody
	// redeployed would lose the registry the moment the password changed: the
	// previous password stays in the account file until the grace ends.
	CredentialRotatedAt time.Time `json:"credentialRotatedAt,omitzero"`
	CredentialGraceEnds time.Time `json:"credentialGraceEnds,omitzero"`
}

type RegistryStorage struct {
	Type base.RegistryStorageType `json:"type,omitempty"`
	// Volume is a cluster-volume setting, for Type volume. The volume decides
	// which node the registry runs on: it already carries that answer, and the
	// placement constraint is derived from it.
	Volume ObjectID `json:"volume,omitzero"`
	// CloudStorage is a cloud-storage setting, for Type s3. Bucket, region and
	// credentials come from it; nothing about S3 is duplicated here.
	CloudStorage ObjectID `json:"cloudStorage,omitzero"`
}

// InUse names the setting this storage actually reads, which is not the same as
// the fields that happen to be filled in: switching between them in the form
// leaves the other one behind.
func (s RegistryStorage) InUse() ObjectID {
	if s.Type == base.RegistryStorageTypeS3 {
		return s.CloudStorage
	}
	return s.Volume
}

type RegistryCleanup struct {
	// Enabled defaults to true in New, so it carries no omitempty: a registry
	// that never prunes is the problem this feature exists to solve.
	Enabled bool                     `json:"enabled"`
	Mode    base.RegistryCleanupMode `json:"mode,omitempty"`
	// KeepLast keeps this many of the newest tags of every repository.
	KeepLast int `json:"keepLast,omitempty"`
	// KeepDays keeps every tag pushed, or pulled, within this many days.
	KeepDays int `json:"keepDays,omitempty"`
}

func (s *RegistrySettings) GetType() base.SettingType {
	return base.SettingTypeRegistry
}

// GetRefObjectIDs reports the volume or the bucket the images are kept in, which
// is what keeps either from being deleted while the registry still uses it.
func (s *RegistrySettings) GetRefObjectIDs() *RefObjectIDs {
	ids := &RefObjectIDs{}
	if inUse := s.Storage.InUse(); inUse.ID != "" {
		ids.RefSettingIDs = append(ids.RefSettingIDs, inUse.ID)
	}
	return ids
}

func (s *RegistrySettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsRegistrySettings() (*RegistrySettings, error) {
	return parseSettingAs[*RegistrySettings](s)
}

func (s *Setting) MustAsRegistrySettings() *RegistrySettings {
	return gofn.Must(s.AsRegistrySettings())
}
