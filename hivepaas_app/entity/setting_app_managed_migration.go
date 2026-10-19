package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

func (s *AppManagedSettings) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppManagedVersion {
		return false, nil
	}
	if setting.Version > CurrentAppManagedVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	// Version 1 is the first, so an older row is one written before versions
	// were set: the data is already in its shape.

	// Version 2 named the engine's state apart from any engine: "postgres" became
	// "database", and a point's or a restore target's "lsn" its "position", which
	// each engine defines. Parsing has already read the old names as the new ones
	// (UnmarshalJSON), and cleared them: writing the data back takes them out of
	// the row.
	if setting.Version < 2 { //nolint:mnd
		s.moveLegacyFields()
	}

	setting.Version = CurrentAppManagedVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}
