package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *CommandPipe) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentCommandPipeVersion {
		return false, nil
	}
	if setting.Version > CurrentCommandPipeVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentCommandPipeVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
