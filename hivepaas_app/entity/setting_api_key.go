package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentAPIKeyVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAPIKey, &apiKeyParser{})

type apiKeyParser struct {
}

func (s *apiKeyParser) New() SettingData {
	return &APIKey{}
}

type APIKey struct {
	KeyID        string              `json:"keyId"`
	SecretKey    HashField           `json:"secretKey"`
	AccessAction *base.AccessActions `json:"accessAction,omitempty"`
}

func (s *APIKey) GetType() base.SettingType {
	return base.SettingTypeAPIKey
}

func (s *APIKey) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *APIKey) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *APIKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAPIKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentAPIKeyVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentAPIKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAPIKey() (*APIKey, error) {
	return parseSettingAs[*APIKey](s)
}

func (s *Setting) MustAsAPIKey() *APIKey {
	return gofn.Must(s.AsAPIKey())
}
