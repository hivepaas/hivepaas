package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *SSHKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSHKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentSSHKeyVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSHKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
