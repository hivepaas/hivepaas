package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *SSHKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSHKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentSSHKeyVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSHKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
