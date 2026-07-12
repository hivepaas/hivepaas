package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentTraefikServiceVersion = 1
)

var _ = registerSettingParser(base.SettingTypeTraefikService, &traefikServiceParser{})

type traefikServiceParser struct {
}

func (s *traefikServiceParser) New() SettingData {
	return &TraefikService{}
}

type TraefikService struct {
	AppSettings TraefikAppSettings `json:"appSettings"`
}

type TraefikAppSettings struct {
	Replicas int `json:"replicas,omitempty"`
}

func (s *TraefikService) GetType() base.SettingType {
	return base.SettingTypeTraefikService
}

func (s *TraefikService) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *TraefikService) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *TraefikService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentTraefikServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentTraefikServiceVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentTraefikServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsTraefikService() (*TraefikService, error) {
	return parseSettingAs[*TraefikService](s)
}

func (s *Setting) MustAsTraefikService() *TraefikService {
	return gofn.Must(s.AsTraefikService())
}
