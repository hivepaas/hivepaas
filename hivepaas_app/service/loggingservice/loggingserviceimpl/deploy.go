package loggingserviceimpl

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// Apply makes the cluster match the stored configuration.
func (s *service) Apply(ctx context.Context, db database.IDB) error {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if setting == nil {
		// Never configured, which is the default: there is nothing to run and
		// nothing to remove.
		return nil
	}

	cfg, err := setting.AsLogging()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if cfg == nil || !cfg.Enabled {
		return s.TearDown(ctx)
	}
	if err := cfg.Decrypt(); err != nil {
		return hperrors.Wrap(err)
	}
	return s.deploy(ctx, cfg)
}

// deploy creates the backend before the collector.
//
// The order is by creation, not readiness: swarm starts tasks asynchronously,
// so the backend may still be starting when the collector comes up. That gap
// costs nothing. vlagent buffers undeliverable logs in a persistent queue under
// -remoteWrite.tmpDataPath - observed opening with persistence enabled - and
// drains it once the backend answers. Waiting on a health check here would add
// a failure mode, a timeout the API holds open, for no lost line.
func (s *service) deploy(ctx context.Context, cfg *entity.Logging) error {
	ingestURL, err := s.deployBackend(ctx, cfg)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if cfg.Collector.Managed {
		if err := s.deployCollector(ctx, cfg, ingestURL); err != nil {
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

// deployCollector runs vlagent on every node, shipping to ingestURL.
func (s *service) deployCollector(ctx context.Context, cfg *entity.Logging, ingestURL string) error {
	spec, err := s.buildCollectSpec(cfg, ingestURL)
	if err != nil {
		return hperrors.Wrap(err)
	}

	collectorImage := ""
	if cfg.Collector.Vlagent != nil {
		collectorImage = cfg.Collector.Vlagent.Image
	}
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
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(s.ensureService(ctx, svcSpec))
}

// deployBackend creates the backend when HivePaaS owns it, and reports where
// the collector should write either way.
func (s *service) deployBackend(ctx context.Context, cfg *entity.Logging) (string, error) {
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
		return "", hperrors.Wrap(loggingservice.ErrNotConfigured)
	}
	if vl.NodeID == "" {
		return "", hperrors.Wrap(loggingservice.ErrBackendNodeMissing)
	}
	if vl.VolumeID == "" {
		// Without a volume the logs live in the container's writable layer and
		// vanish on the next restart, silently.
		return "", hperrors.Wrap(loggingservice.ErrVolumeMissing)
	}

	deployer, err := logging.NewDeployer(
		logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{
			Image:               vl.Image,
			DataVolumeName:      vl.VolumeID,
			Retention:           vl.Retention.ToDuration(),
			MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
		}},
	)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	rt, err := deployer.RuntimeSpec()
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	svcSpec, err := toSwarmServiceSpec(rt, swarmSpecOpts{
		Name:   ServiceNameBackend,
		NodeID: vl.NodeID,
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	if err := s.ensureService(ctx, svcSpec); err != nil {
		return "", hperrors.Wrap(err)
	}

	return fmt.Sprintf("http://%s:%d%s",
		ServiceNameBackend, victorialogs.DefaultHTTPPort, victorialogs.IngestPath), nil
}

// TearDown removes what Apply created, in the reverse order, and leaves the
// data volume alone.
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
func (s *service) Status(ctx context.Context, db database.IDB) (*loggingservice.Status, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeLogging, true)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	out := &loggingservice.Status{}
	if setting == nil {
		out.ExcludedApps, err = s.listExcludedApps(ctx, db)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return out, nil
	}
	cfg, err := setting.AsLogging()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if cfg != nil {
		out.Enabled = cfg.Enabled
	}

	// Listed whether or not logging is on: the list is most useful before it is
	// turned on, when there is still time to switch the apps it names.
	out.ExcludedApps, err = s.listExcludedApps(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	backend, inspectErr := s.clusterService.ServiceInspect(ctx, ServiceNameBackend, true)
	if svc := backend; inspectErr == nil && svc != nil {
		out.BackendServiceID = svc.ID
		out.BackendReady = true
	}
	collector, inspectErr := s.clusterService.ServiceInspect(ctx, ServiceNameCollector, true)
	if svc := collector; inspectErr == nil && svc != nil {
		out.CollectorService = svc.ID
	}
	return out, nil
}
