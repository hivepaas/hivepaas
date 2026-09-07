package entity

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// hivePaaSServiceVersionProxyHops is the version that made the forwarded-header
// depth an explicit setting instead of a constant in the Traefik label builder.
const hivePaaSServiceVersionProxyHops = 2

func (s *HivePaaSService) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentHivePaaSServiceVersion {
		return false, nil
	}
	if setting.Version > CurrentHivePaaSServiceVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}

	if setting.Version < hivePaaSServiceVersionProxyHops {
		s.migrateProxyHops()
	}

	setting.Version = CurrentHivePaaSServiceVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

// migrateProxyHops carries the old hard-coded depth forward.
//
// Only for installs that already declared a proxy. Writing it everywhere would
// tell installs with nothing in front to read a forwarded header, which is the
// one configuration that must never be guessed: with no proxy that header is
// written by the caller, so every request would identify as whoever it likes.
func (s *HivePaaSService) migrateProxyHops() {
	if s.ProxySettings.HasProxy() && s.ProxySettings.ProxyHops <= 0 {
		s.ProxySettings.ProxyHops = LegacyProxyHops
	}
}
