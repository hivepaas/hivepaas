package loggingserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// Apply makes the cluster match the stored configuration.
func (s *service) Apply(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
) (_ *loggingservice.SettingApplyResp, err error) {
	setting := req.Setting
	if setting == nil {
		setting, err = s.loadSettings(ctx, db, false)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if setting == nil {
			// Never configured, which is the default: there is nothing to run and nothing to remove.
			return nil, nil
		}
	}

	cfg, err := setting.AsLoggingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg == nil || !cfg.Enabled {
		return nil, s.TearDown(ctx)
	}
	if err := cfg.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return nil, s.deploy(ctx, db, cfg)
}

// deploy creates the backend before the collector.
//
// The order is by creation, not readiness: swarm starts tasks asynchronously,
// so the backend may still be starting when the collector comes up. That gap
// costs nothing. vlagent buffers undeliverable logs in a persistent queue under
// -remoteWrite.tmpDataPath - observed opening with persistence enabled - and
// drains it once the backend answers. Waiting on a health check here would add
// a failure mode, a timeout the API holds open, for no lost line.
func (s *service) deploy(ctx context.Context, db database.IDB, cfg *entity.LoggingSettings) error {
	logNet, err := s.ensureLoggingNetwork(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}

	ingestURL, err := s.deployBackend(ctx, db, cfg, logNet)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if cfg.Collector.Managed {
		if err := s.deployCollector(ctx, cfg, ingestURL, logNet); err != nil {
			return hperrors.Wrap(err)
		}
	} else {
		// The collector is somebody else's now. One HivePaaS ran before would
		// otherwise keep shipping alongside theirs.
		if err := s.removeService(ctx, ServiceNameCollector); err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Removed last, once no collector HivePaaS runs points at it any more. The
	// data volume stays: see TearDown.
	if !cfg.Backend.Managed {
		if err := s.removeService(ctx, ServiceNameBackend); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// deployCollector runs vlagent on every node, shipping to ingestURL over the
// logging network.
func (s *service) deployCollector(ctx context.Context, cfg *entity.LoggingSettings, ingestURL, logNet string) error {
	spec, err := s.buildCollectSpec(cfg, ingestURL)
	if err != nil {
		return hperrors.Wrap(err)
	}

	_, collectorImage := releaseImages()
	collector, err := logging.NewCollector(
		logging.CollectorType(cfg.Collector.Type),
		&logging.CollectorConfig{Vlagent: &vlagent.Config{Image: collectorImage}},
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	rt, err := collector.Configure(spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	svcSpec, err := toSwarmServiceSpec(rt, swarmSpecOpts{
		Name: ServiceNameCollector,
		// One task per node; this is the only global service HivePaaS creates.
		Global: true,
		// Only the logging network. The API's private network holds the
		// database, and this runs on every node.
		Networks: []string{logNet},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(s.ensureService(ctx, svcSpec))
}

// deployBackend creates the backend when HivePaaS owns it, and reports where
// the collector should write either way.
func (s *service) deployBackend(
	ctx context.Context,
	db database.IDB,
	cfg *entity.LoggingSettings,
	logNet string,
) (string, error) {
	if !cfg.Backend.Managed {
		ep, err := toEndpoint(cfg.Backend.Ingest)
		if err != nil {
			return "", hperrors.Wrap(err)
		}
		if ep.URL == "" {
			return "", hperrors.Wrap(logging.ErrIngestEndpointRequired)
		}
		return ep.URL, nil
	}

	vl := cfg.Backend.VictoriaLogs
	if vl == nil {
		return "", hperrors.Wrap(hperrors.ErrLoggingNotConfigured)
	}
	if vl.Volume.ID == "" {
		// Without a volume the logs live in the container's writable layer and
		// vanish on the next restart, silently.
		return "", hperrors.Wrap(hperrors.ErrLoggingVolumeMissing)
	}

	constraint, err := s.backendPlacement(ctx, db, vl.Volume.ID)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	backendImage, _ := releaseImages()
	deployer, err := logging.NewDeployer(
		logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{
			Image:               backendImage,
			DataVolumeName:      vl.Volume.ID,
			DataSubpath:         vl.VolumeSubpath,
			Retention:           vl.Retention.ToDuration(),
			MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
			Resources:           toLoggingResources(vl),
		}},
	)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	rt, err := deployer.RuntimeSpec()
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	// The collector reaches it over the logging network, the API over the
	// stack's internal one.
	if err := s.checkInternalNetwork(ctx); err != nil {
		return "", hperrors.Wrap(err)
	}

	svcSpec, err := toSwarmServiceSpec(rt, swarmSpecOpts{
		Name:       ServiceNameBackend,
		Constraint: constraint,
		Resources:  rt.Resources,
		Networks:   []string{logNet, base.NetworkHivepaasLocal},
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if err := s.ensureService(ctx, svcSpec); err != nil {
		return "", hperrors.Wrap(err)
	}

	return backendBaseURL() + victorialogs.IngestPath, nil
}

// backendPlacement is where the backend must run, taken from the volume it
// writes to rather than asked for separately.
//
// One answer cannot disagree with itself: a placement chosen apart from the
// volume can name a node the volume is not on, and the backend then starts on
// an empty directory. A volume pinned to nowhere returns no constraint, which
// leaves swarm free to place - and free to move - the backend.
func (s *service) backendPlacement(ctx context.Context, db database.IDB, volumeID string) (string, error) {
	setting, err := s.settingRepo.GetByID(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeClusterVolume, volumeID, true)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if setting == nil {
		return "", hperrors.Wrap(hperrors.ErrLoggingVolumeMissing)
	}
	vol, err := setting.AsClusterVolume()
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	constraint, conflict := placementservice.VolumePinConstraint([]placementservice.VolumePin{{
		VolumeName: setting.Name,
		NodeID:     vol.NodeID,
		NodeLabel:  vol.NodeLabel,
	}})
	if conflict != nil {
		// One volume cannot conflict with itself; the signature allows many.
		return "", hperrors.Wrap(hperrors.ErrLoggingDeployFailed).WithExtraDetail("%s", conflict.Error())
	}
	return constraint, nil
}

func toLoggingResources(vl *entity.LoggingVictoriaLogs) logging.Resources {
	return logging.Resources{CPULimit: vl.CPULimit, MemoryLimit: vl.MemoryLimit.Bytes()}
}

// TearDown removes what Apply created, in the reverse order, and leaves the
// data volume alone.
//
// The logging network stays too: an empty overlay costs nothing, and removing
// it while a task is still draining fails.
//
// The volume is deliberately kept. Turning logging off, or moving to a backend
// HivePaaS does not run, must not destroy what has been collected - for a plain
// local volume the data is inside the volume and removal is unrecoverable.
// Deleting it is a separate, explicit action through the volume API.
func (s *service) TearDown(ctx context.Context) error {
	var errs error
	for _, name := range []string{ServiceNameCollector, ServiceNameBackend} {
		if err := s.removeService(ctx, name); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if errs != nil {
		return hperrors.Wrap(errs)
	}
	return nil
}

// ensureService creates the service, or updates it in place when one by that
// name already exists.
//
// Apply runs every time the settings are saved, so creating unconditionally
// would fail the second save on a name clash. Updating in place also means a
// changed retention or a new forward rolls the running service rather than
// removing it and starting over.
func (s *service) ensureService(ctx context.Context, spec *swarm.ServiceSpec) error {
	existing, err := s.dockerManager.ServiceInspect(ctx, spec.Name)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if err != nil || existing == nil {
		_, err = s.dockerManager.ServiceCreate(ctx, spec)
		return hperrors.Wrap(err)
	}
	_, err = s.dockerManager.ServiceUpdate(ctx, existing.Service.ID, &existing.Service.Version, spec)
	return hperrors.Wrap(err)
}

// removeService removes the service by name, treating one that is not there as
// already removed.
func (s *service) removeService(ctx context.Context, name string) error {
	_, err := s.dockerManager.ServiceRemove(ctx, name)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	return nil
}

// Status reports what is running.
func (s *service) Status(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
) (status *loggingservice.Status, err error) {
	if setting == nil {
		setting, err = s.loadSettings(ctx, db, false)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	status = &loggingservice.Status{}
	if setting == nil {
		return status, nil
	}

	loggingSettings, err := setting.AsLoggingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	status.Enabled = loggingSettings.Enabled
	if !status.Enabled {
		return status, nil
	}

	collector, inspectErr := s.clusterService.ServiceInspect(ctx, ServiceNameCollector, true)
	if svc := collector; inspectErr == nil && svc != nil {
		status.CollectorServiceID = svc.ID
		switch {
		case svc.Spec.Mode.Replicated != nil:
			if svc.Spec.Mode.Replicated.Replicas != nil && *svc.Spec.Mode.Replicated.Replicas > 0 {
				status.CollectorReady = true
			}
		case svc.Spec.Mode.Global != nil:
			status.CollectorReady = true
		}
	}

	backend, inspectErr := s.clusterService.ServiceInspect(ctx, ServiceNameBackend, true)
	if svc := backend; inspectErr == nil && svc != nil {
		status.BackendServiceID = svc.ID
		switch {
		case svc.Spec.Mode.Replicated != nil:
			if svc.Spec.Mode.Replicated.Replicas != nil && *svc.Spec.Mode.Replicated.Replicas > 0 {
				status.BackendReady = true
			}
		case svc.Spec.Mode.Global != nil:
			status.BackendReady = true
		}
	}

	return status, nil
}
