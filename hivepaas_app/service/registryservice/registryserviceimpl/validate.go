package registryserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// validateSettings refuses a configuration that could not produce a working
// registry, and one that would throw away a working one.
//
// current is what is stored today, nil when nothing is. It is only needed for the
// questions that are about a change rather than about a value.
//
// provisioned says the registry's app exists right now. It is asked rather than
// read from current.AppID: that id outlives the app when somebody deletes it in
// the app screen - which is exactly how the spec says a registry is removed - and
// trusting it made a deleted registry impossible to provision again.
func validateSettings(cfg, current *entity.RegistrySettings, provisioned bool) error {
	if cfg == nil {
		return hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}
	// Nothing below is deployed while it is off, so nothing below has to be
	// answerable yet.
	if !cfg.Enabled {
		return nil
	}

	if cfg.Domain == "" {
		return invalid("A domain is required: images are named after it, and Docker only " +
			"speaks to a registry over HTTPS at a name.")
	}
	switch cfg.Storage.Type {
	case base.RegistryStorageTypeVolume:
		if cfg.Storage.Volume.ID == "" {
			return invalid("Choose the volume the images are kept on.")
		}
	case base.RegistryStorageTypeS3:
		if cfg.Storage.CloudStorage.ID == "" {
			return invalid("Choose the cloud storage the images are kept in.")
		}
	default:
		return invalid("Unknown storage type %q.", cfg.Storage.Type)
	}
	if cfg.Cleanup.Enabled {
		if cfg.Cleanup.KeepLast < 1 {
			return invalid("Builds to keep must be at least one.")
		}
		if cfg.Cleanup.KeepDays < 1 {
			return invalid("Days to keep must be at least one.")
		}
	}

	// Once the app exists the storage is what holds its images. Nothing copies
	// them anywhere, so changing it is refused rather than obeyed.
	if provisioned && current != nil {
		if cfg.Storage.Type != current.Storage.Type ||
			cfg.Storage.InUse().ID != current.Storage.InUse().ID {
			return hperrors.Wrap(hperrors.ErrRegistryStorageImmutable).WithExtraDetail(
				"It holds images, and nothing copies them from one store to the other. " +
					"To move them, delete the registry app and provision a new registry.")
		}
	}
	return nil
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).WithExtraDetail(format, args...)
}
