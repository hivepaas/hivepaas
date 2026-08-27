package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *GithubApp) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentGithubAppVersion {
		return false, nil
	}
	if setting.Version > CurrentGithubAppVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentGithubAppVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
