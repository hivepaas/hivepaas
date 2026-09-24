package systemappservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

// CurrentRelease is the release this binary runs, which names the images of the
// system apps.
//
// The images come from the release rather than from any stored setting because
// the system updater is what moves them. The release the updater applies and the
// release compiled into the binary it installs are the same one, so the two agree
// - as long as release.json and base.ReleaseInfo are changed together, which is
// the contract stated there.
func CurrentRelease() *base.ReleaseInfo {
	// config.Current() is nil until a config has been loaded, which is the case
	// for tests exercising a provisioning path on its own. Stable is the right
	// answer to "no idea which channel this is".
	if cfg := config.Current(); cfg != nil && cfg.IsBetaEnv() {
		return base.BetaVersion
	}
	return base.StableVersion
}
