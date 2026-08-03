package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *RepoWebhook) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentRepoWebhookVersion {
		return false, nil
	}
	if setting.Version > CurrentRepoWebhookVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentRepoWebhookVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
