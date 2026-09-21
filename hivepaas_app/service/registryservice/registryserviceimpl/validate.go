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
func validateSettings(cfg, current *entity.RegistrySettings) error {
	if cfg == nil {
		return hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}
	// Nothing below is deployed while it is off, so nothing below has to be
	// answerable yet.
	if !cfg.Enabled {
		return nil
	}

	if cfg.Domain == "" {
		return invalid("a domain is required: images are named after it, and docker only " +
			"speaks to a registry over HTTPS at a name")
	}
	switch cfg.Storage.Type {
	case base.RegistryStorageTypeVolume:
		if cfg.Storage.Volume.ID == "" {
			return invalid("choose the volume the images are kept on")
		}
	case base.RegistryStorageTypeS3:
		if cfg.Storage.CloudStorage.ID == "" {
			return invalid("choose the cloud storage the images are kept in")
		}
	default:
		return invalid("unknown storage type %q", cfg.Storage.Type)
	}
	if cfg.Cleanup.Enabled {
		if cfg.Cleanup.KeepLast < 1 {
			return invalid("keepLast must keep at least one build of every app")
		}
		if cfg.Cleanup.KeepDays < 1 {
			return invalid("keepDays must be at least one day")
		}
	}
	if cfg.MemoryLimit < entity.MinRegistryMemoryLimit {
		return invalid("the memory limit must be at least 256MB")
	}

	// Once the app exists the storage is what holds its images. Nothing copies
	// them anywhere, so changing it is refused rather than obeyed.
	if current != nil && current.AppID != "" {
		if cfg.Storage.Type != current.Storage.Type ||
			cfg.Storage.InUse().ID != current.Storage.InUse().ID {
			return hperrors.Wrap(hperrors.ErrRegistryStorageImmutable).WithExtraDetail(
				"the registry's storage cannot be changed once it holds images: " +
					"provision a new registry instead")
		}
	}
	return nil
}

func invalid(format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).WithExtraDetail(format, args...)
}
