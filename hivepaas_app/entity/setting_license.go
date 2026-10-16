package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentLicenseSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeLicense, &licenseSettingsParser{})

type licenseSettingsParser struct {
}

// New is no key at all, which is what an installation runs on until one is
// pasted: the free plan.
func (s *licenseSettingsParser) New() SettingData {
	return &LicenseSettings{}
}

// LicenseSettings is the license key this installation runs on. It is global,
// and there is one. See docs/superpowers/specs/2026-10-16-licensing-design.md.
type LicenseSettings struct {
	// Key is the license key as pasted, normalized. It is not a secret, but it
	// carries the customer's email, so it is never logged and never exported.
	Key string `json:"key"`
	// MovedAt is when online refresh learned that the key was moved to another
	// installation; zero while it was not. Phase 2 writes it; phase 1 only reads it.
	MovedAt time.Time `json:"movedAt,omitzero"`
}

func (s *LicenseSettings) GetType() base.SettingType {
	return base.SettingTypeLicense
}

func (s *LicenseSettings) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *LicenseSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsLicenseSettings() (*LicenseSettings, error) {
	return parseSettingAs[*LicenseSettings](s)
}

func (s *Setting) MustAsLicenseSettings() *LicenseSettings {
	return gofn.Must(s.AsLicenseSettings())
}
