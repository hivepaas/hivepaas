package entity

import (
	"github.com/tiendc/gofn"

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

func (s *TraefikService) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsTraefikService() (*TraefikService, error) {
	return parseSettingAs[*TraefikService](s)
}

func (s *Setting) MustAsTraefikService() *TraefikService {
	return gofn.Must(s.AsTraefikService())
}
