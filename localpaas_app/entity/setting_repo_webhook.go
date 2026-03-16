package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
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

func (s *RepoWebhook) MustDecrypt() *RepoWebhook {
	return s
}

func (s *RepoWebhook) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentRepoWebhookVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRepoWebhookVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsRepoWebhook() (*RepoWebhook, error) {
	return parseSettingAs[*RepoWebhook](s)
}

func (s *Setting) MustAsRepoWebhook() *RepoWebhook {
	return gofn.Must(s.AsRepoWebhook())
}
