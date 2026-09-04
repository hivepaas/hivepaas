package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
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
	Secret EncryptedField   `json:"secret"`
}

func (s *RepoWebhook) GetType() base.SettingType {
	return base.SettingTypeRepoWebhook
}

func (s *RepoWebhook) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *RepoWebhook) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *RepoWebhook) Decrypt() error {
	_, err := s.Secret.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *Setting) AsRepoWebhook() (*RepoWebhook, error) {
	// Github-app setting can be parsed as RepoWebhook
	if s.Type == base.SettingTypeGithubApp {
		ghApp, err := s.AsGithubApp()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return ghApp.ConvertAsRepoWebhook(), nil
	}
	return parseSettingAs[*RepoWebhook](s)
}

func (s *Setting) MustAsRepoWebhook() *RepoWebhook {
	return gofn.Must(s.AsRepoWebhook())
}
