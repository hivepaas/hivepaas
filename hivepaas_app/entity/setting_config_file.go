package entity

import (
	"encoding/base64"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

const (
	CurrentConfigFileVersion = 1
)

var _ = registerSettingParser(base.SettingTypeConfigFile, &configFileParser{})

type configFileParser struct {
}

func (s *configFileParser) New() SettingData {
	return &ConfigFile{}
}

type ConfigFile struct {
	Name     string          `json:"name"`
	Content  string          `json:"content"`
	Base64   bool            `json:"base64,omitempty"`
	SwarmRef *SwarmConfigRef `json:"swarmRef,omitempty"`
}

type SwarmConfigRef struct {
	File       *SwarmRefFileTarget `json:"file"`
	ConfigID   string              `json:"configId"`
	ConfigName string              `json:"configName"`
}

func (s *ConfigFile) GetType() base.SettingType {
	return base.SettingTypeConfigFile
}

func (s *ConfigFile) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *ConfigFile) ContentAsBytes() []byte {
	if s.Base64 {
		return gofn.Must(base64.StdEncoding.DecodeString(s.Content))
	}
	return reflectutil.UnsafeStrToBytes(s.Content)
}

func (s *ConfigFile) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *ConfigFile) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentConfigFileVersion {
		return false, nil
	}
	if setting.Version > CurrentConfigFileVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentConfigFileVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsConfigFile() (*ConfigFile, error) {
	return parseSettingAs[*ConfigFile](s)
}

func (s *Setting) MustAsConfigFile() *ConfigFile {
	return gofn.Must(s.AsConfigFile())
}
