package clustercleanupserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func (s *service) cleanupCluster(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	clusterCleanup := data.CleanupSettings
	if !clusterCleanup.Enabled {
		return nil
	}

	if e := s.cleanupContainersAndServices(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	if e := s.cleanupImages(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	if e := s.cleanupVolumes(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	if e := s.cleanupNetworks(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	if e := s.cleanupBuildCache(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	return apperrors.Wrap(err)
}

func (s *service) cleanupContainersAndServices(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	objectRetention := data.CleanupSettings.GeneralRetention.ToDuration()

	if data.CleanupContainers != base.CleanupFlagFalse && data.CleanupSettings.PruneContainers {
		resp, e := s.dockerManager.ContainerPrune(ctx, objectRetention)
		if e != nil {
			data.Output.ContainersPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.Output.ContainersDeleted = len(report.ContainersDeleted)
			data.Output.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	// Always cleanup orphaned temporary batch containers (hivepaas-cont-) older than grace period
	if e := s.cleanupTempContainers(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	// Always cleanup orphaned temporary batch services (hivepaas-svc-) older than grace period
	if e := s.cleanupTempServices(ctx, data); e != nil {
		err = errors.Join(err, e)
	}

	return apperrors.Wrap(err)
}

func (s *service) cleanupImages(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	objectRetention := data.CleanupSettings.GeneralRetention.ToDuration()

	if data.CleanupImages != base.CleanupFlagFalse && data.CleanupSettings.PruneImages {
		resp, e := s.dockerManager.ImagePrune(ctx, false, objectRetention)
		if e != nil {
			data.Output.ImagesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.Output.ImagesDeleted = len(report.ImagesDeleted)
			data.Output.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	return apperrors.Wrap(err)
}

func (s *service) cleanupVolumes(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	if data.CleanupVolumes != base.CleanupFlagFalse && data.CleanupSettings.PruneVolumes {
		resp, e := s.dockerManager.VolumePrune(ctx, true)
		if e != nil {
			data.Output.VolumesPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.Output.VolumesDeleted = len(report.VolumesDeleted)
			data.Output.SpaceReclaimed += report.SpaceReclaimed
		}
	}

	return apperrors.Wrap(err)
}

func (s *service) cleanupNetworks(
	ctx context.Context,
	data *clusterCleanupData,
) (err error) {
	objectRetention := data.CleanupSettings.GeneralRetention.ToDuration()

	if data.CleanupNetworks != base.CleanupFlagFalse && data.CleanupSettings.PruneNetworks {
		resp, e := s.dockerManager.NetworkPrune(ctx, objectRetention)
		if e != nil {
			data.Output.NetworksPruneError = e.Error()
			err = errors.Join(err, e)
		} else {
			report := &resp.Report
			data.Output.NetworksDeleted = len(report.NetworksDeleted)
		}
	}

	return apperrors.Wrap(err)
}
