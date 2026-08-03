package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func (s *CommandPipe) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCommandPipeVersion {
		return false, nil
	}
	if setting.Version > CurrentCommandPipeVersion {
		return false, apperrors.Wrap(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCommandPipeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
