package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAppCloneSettingsVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppClone, &appCloneSettingsParser{})

type appCloneSettingsParser struct {
}

func (s *appCloneSettingsParser) New() SettingData {
	return &AppCloneSettings{}
}

type AppCloneSettings struct {
	TargetName     string         `json:"targetName,omitempty"`
	TargetEnv      string         `json:"targetEnv,omitempty"`
	TargetStatus   base.AppStatus `json:"targetStatus,omitempty"`
	TargetReplicas int            `json:"targetReplicas,omitempty"`

	CloneDeploymentSettings bool                             `json:"cloneDeploymentSettings"`
	CloneRoutingSettings    bool                             `json:"cloneRoutingSettings"`
	CloneRoutingDomains     []*AppCloneRoutingDomainSettings `json:"cloneRoutingDomains,omitempty"`

	CloneVolumes    bool     `json:"cloneVolumes"`
	CloneVolumeData bool     `json:"cloneVolumeData"`
	LiveVolumeClone bool     `json:"liveVolumeClone"`
	IncludedVolumes []string `json:"includedVolumes,omitempty"` // ids or names
	ExcludedVolumes []string `json:"excludedVolumes,omitempty"` // ids or names

	CloneEnvVars      bool `json:"cloneEnvVars"`
	CloneSecrets      bool `json:"cloneSecrets"`
	CloneConfigFiles  bool `json:"cloneConfigFiles"`
	ClonePeriodicJobs bool `json:"clonePeriodicJobs"`
	CloneSchedJobs    bool `json:"cloneSchedJobs"`

	// CommandPipes commands to run after cloning
	CommandPipes ObjectIDSlice `json:"commandPipes,omitempty"`

	Notification *BaseEventNotification `json:"notification,omitempty"`
}

type AppCloneRoutingDomainSettings struct {
	SourceDomain  string   `json:"sourceDomain"`
	TargetDomain  string   `json:"targetDomain"`
	SourceSSLCert ObjectID `json:"sourceSslCert,omitzero"`
	TargetSSLCert ObjectID `json:"targetSslCert,omitzero"`
}

func (s *AppCloneSettings) GetType() base.SettingType {
	return base.SettingTypeAppClone
}

func (s *AppCloneSettings) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	for _, domain := range s.CloneRoutingDomains {
		if domain.SourceSSLCert.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, domain.SourceSSLCert.ID)
		}
		if domain.TargetSSLCert.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, domain.TargetSSLCert.ID)
		}
	}
	for _, cmdPipe := range s.CommandPipes {
		if cmdPipe.ID != "" {
			refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, cmdPipe.ID)
		}
	}
	refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	return refIDs
}

func (s *AppCloneSettings) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppCloneSettings() (*AppCloneSettings, error) {
	return parseSettingAs[*AppCloneSettings](s)
}

func (s *Setting) MustAsAppCloneSettings() *AppCloneSettings {
	return gofn.Must(s.AsAppCloneSettings())
}
