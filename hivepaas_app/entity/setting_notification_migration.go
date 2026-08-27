package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *Notification) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentNotificationVersion {
		return false, nil
	}
	if setting.Version > CurrentNotificationVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentNotificationVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
