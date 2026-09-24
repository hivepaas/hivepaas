package loggingserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Validate refuses a volume the backend could not use, and a change that would
// throw away a running backend's store.
//
// What a configuration must hold regardless of what exists - a volume at all, a
// source, an ingest endpoint - is the usecase's to check; this is the part that
// needs the database.
func (s *service) Validate(
	ctx context.Context,
	db database.IDB,
	next, current *entity.LoggingSettings,
) error {
	if next == nil || !next.Enabled || !next.Backend.Managed ||
		next.Backend.VictoriaLogs == nil || next.Backend.VictoriaLogs.Volume.ID == "" {
		return nil
	}
	volumeID := next.Backend.VictoriaLogs.Volume.ID

	backend, err := s.systemAppService.LoadApp(ctx, db, backendAppKey)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Once the backend exists the volume is what holds its logs. Nothing copies
	// them anywhere, so changing it is refused rather than obeyed.
	if backend != nil {
		if current != nil && current.Backend.VictoriaLogs != nil &&
			current.Backend.VictoriaLogs.Volume.ID != "" && current.Backend.VictoriaLogs.Volume.ID != volumeID {
			return hperrors.Wrap(hperrors.ErrLoggingVolumeImmutable).WithExtraDetail(
				"It holds the logs collected so far, and nothing copies them to another volume. " +
					"To move them, switch logging off with its apps removed, then on again with the new volume.")
		}
		return nil
	}

	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeClusterVolume, volumeID, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		return hperrors.Wrap(hperrors.ErrLoggingSettingsInvalid).
			WithExtraDetail("The volume the logs are to be kept on no longer exists.")
	}
	// A volume reaches an app only when it is inheritable, and the backend's app
	// is in a project of its own. Without this the failure lands inside the build
	// of its mounts as "Volume not found", which says nothing about what to do.
	if !setting.Inheritable {
		return hperrors.Wrap(hperrors.ErrLoggingSettingsInvalid).WithExtraDetail(
			"The volume %q is not shared with apps. Edit it in Cluster > Volumes and make it "+
				"available to apps, or choose one that already is.", setting.Name)
	}
	return nil
}
