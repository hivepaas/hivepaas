package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentLicenseSettingsVersion = 1

	// LicenseSettingName is the license setting's name, whoever writes it first.
	LicenseSettingName = "License"
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
	// installation; zero while it was not. Installing a key clears it.
	MovedAt time.Time `json:"movedAt,omitzero"`
	// Refresh is the daily check with the portal for a newer key, and how the
	// last one went. Installing a key keeps it: it is about this installation,
	// not about the key.
	Refresh LicenseRefresh `json:"refresh,omitzero"`
}

// LicenseRefresh is the daily check with the license portal and its last
// outcome. Only a check that asked the portal, or failed trying, is recorded:
// one with no key to ask with, or with the daily check switched off, is not.
type LicenseRefresh struct {
	// Schedule is the daily job's, kept here as every daily system job keeps its
	// own, so moving the jobs to a new timezone moves it too.
	Schedule SchedJobSchedule `json:"schedule,omitzero"`
	// LastCheckedAt is when the portal was last asked, whatever came of it.
	LastCheckedAt time.Time                 `json:"lastCheckedAt,omitzero"`
	LastResult    base.LicenseRefreshResult `json:"lastResult,omitempty"`
	// LastError says why the last check failed; empty when it did not. It holds
	// neither the key nor anything the key says.
	LastError string `json:"lastError,omitempty"`
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
