package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	CurrentRepoWebhookVersion = 1
)

var _ = registerSettingParser(base.SettingTypeRepoWebhook, &repoWebhookParser{})

type repoWebhookParser struct {
}

func (s *repoWebhookParser) New() SettingData {
	return &RepoWebhook{}
}

type RepoWebhook struct {
	Kind   base.WebhookKind `json:"kind"`
	Secret string           `json:"secret"`
}

func (s *RepoWebhook) GetType() base.SettingType {
	return base.SettingTypeRepoWebhook
}

func (s *RepoWebhook) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *RepoWebhook) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *RepoWebhook) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRepoWebhookVersion {
		return false, nil
	}
	if setting.Version > CurrentRepoWebhookVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRepoWebhookVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *RepoWebhook) Decrypt() error {
	return nil
}

func (s *Setting) AsRepoWebhook() (*RepoWebhook, error) {
	// Github-app setting can be parsed as RepoWebhook
	if s.Type == base.SettingTypeGithubApp {
		ghApp, err := s.AsGithubApp()
		if err != nil {
			return nil, apperrors.New(err)
		}
		return ghApp.ConvertAsRepoWebhook(), nil
	}
	return parseSettingAs[*RepoWebhook](s)
}

func (s *Setting) MustAsRepoWebhook() *RepoWebhook {
	return gofn.Must(s.AsRepoWebhook())
}
