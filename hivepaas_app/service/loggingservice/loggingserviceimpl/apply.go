package loggingserviceimpl

import (
	"context"
	"errors"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

const (
	backendAppName   = "Logging backend"
	collectorAppName = "Logging collector"

	backendAppNote = "The log store HivePaaS runs for itself (VictoriaLogs). " +
		"It is configured in system settings, under Logging."
	collectorAppNote = "The log collector HivePaaS runs on every node (vlagent). " +
		"It is configured in system settings, under Logging."
)

// Apply makes the cluster match the stored configuration.
//
// The backend comes before the collector and goes after it: the collector needs
// somewhere to write, and should not ship into a store that is gone. The order
// is by creation, not readiness - vlagent buffers what it cannot deliver in a
// persistent queue and drains it once the backend answers, so waiting on a
// health check here would add a timeout for no lost line.
func (s *service) Apply(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
) (_ *loggingservice.SettingApplyResp, err error) {
	resp := &loggingservice.SettingApplyResp{}
	setting := req.Setting
	if setting == nil {
		setting, err = s.loadSettings(ctx, db, false)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if setting == nil {
			// Never configured, which is the default: there is nothing to run.
			return resp, nil
		}
	}
	cfg, err := setting.AsLoggingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = cfg.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}

	backend, err := s.systemAppService.LoadApp(ctx, db, backendAppKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	collector, err := s.systemAppService.LoadApp(ctx, db, collectorAppKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = s.removeAppsNoLongerRun(ctx, db, req, cfg, &backend, &collector, resp); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// From here on the response goes back with the error too: it carries the
	// cleanup of whatever provisioning already created in docker.
	if cfg.Enabled {
		ingest, applyErr := s.applyBackend(ctx, db, req, cfg, &backend, resp)
		if applyErr != nil {
			return resp, hperrors.Wrap(applyErr)
		}
		if cfg.Collector.Managed {
			if applyErr = s.applyCollector(ctx, db, req, cfg, ingest, &collector, resp); applyErr != nil {
				return resp, hperrors.Wrap(applyErr)
			}
		}
	}

	if err = s.rememberApps(ctx, db, setting, cfg, backend, collector); err != nil {
		return resp, hperrors.Wrap(err)
	}
	return resp, nil
}

// removeAppsNoLongerRun takes down the apps HivePaaS stops running - both when
// logging is switched off, one when it is handed to a system somebody else
// runs - once the request has confirmed it. The collector goes first, so that
// nothing ships into a store that is going away.
func (s *service) removeAppsNoLongerRun(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
	cfg *entity.LoggingSettings,
	backend, collector **entity.App,
	resp *loggingservice.SettingApplyResp,
) error {
	var removing []**entity.App
	if *collector != nil && (!cfg.Enabled || !cfg.Collector.Managed) {
		removing = append(removing, collector)
	}
	if *backend != nil && (!cfg.Enabled || !cfg.Backend.Managed) {
		removing = append(removing, backend)
	}
	if len(removing) == 0 {
		return nil
	}
	if !req.RemoveApp {
		names := make([]string, 0, len(removing))
		for _, app := range removing {
			names = append(names, (*app).Name)
		}
		return hperrors.Wrap(hperrors.ErrLoggingAppStillRunning).WithExtraDetail(
			"This takes down %s. Confirm the removal, and say whether the stored logs should go with it.",
			strings.Join(names, " and "))
	}
	for _, app := range removing {
		// The stored logs are the backend's; the collector keeps nothing worth it.
		removeStorage := req.RemoveStorage && app == backend
		s.logger.Info("removing a logging app", "app", (*app).ID, "key", (*app).Key,
			"removeStorage", removeStorage)
		if err := s.systemAppService.Remove(ctx, db, *app, removeStorage); err != nil {
			return hperrors.Wrap(err)
		}
		*app = nil
		resp.RemovedApps = true
	}
	return nil
}

// applyBackend provisions or reconciles the backend HivePaaS runs, and says
// where the collector should write - to it, or to the store somebody else runs.
func (s *service) applyBackend(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
	cfg *entity.LoggingSettings,
	backend **entity.App,
	resp *loggingservice.SettingApplyResp,
) (*logging.Endpoint, error) {
	if !cfg.Backend.Managed {
		ingest, err := toEndpoint(cfg.Backend.Ingest)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if ingest == nil || ingest.URL == "" {
			return nil, hperrors.Wrap(logging.ErrIngestEndpointRequired)
		}
		// Whole, credentials and all: a store somebody else runs usually wants them.
		return ingest, nil
	}

	vl := cfg.Backend.VictoriaLogs
	if vl == nil {
		return nil, hperrors.Wrap(hperrors.ErrLoggingNotConfigured)
	}
	backendImage, _ := releaseImages()
	deployer, err := logging.NewDeployer(logging.BackendType(cfg.Backend.Type),
		&logging.BackendConfig{VictoriaLogs: &victorialogs.Config{
			Image:               backendImage,
			DataVolume:          vl.Volume.ID,
			Retention:           vl.Retention.ToDuration(),
			MaxDiskUsagePercent: vl.MaxDiskUsagePercent,
		}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	rt, err := deployer.RuntimeSpec()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ingest := &logging.Endpoint{URL: backendBaseURL() + victorialogs.IngestPath}

	if *backend != nil {
		if err = s.redeploy(ctx, db, req, *backend, rt, resp); err != nil {
			return nil, hperrors.Wrap(err)
		}
		if req.BackendResources != nil {
			if err = s.applyBackendResources(ctx, *backend, *req.BackendResources); err != nil {
				return nil, hperrors.Wrap(err)
			}
		}
		return ingest, nil
	}

	// The API queries the backend over the stack's internal network, so a
	// backend it cannot reach is refused rather than deployed.
	if err = s.checkInternalNetwork(ctx); err != nil {
		return nil, hperrors.Wrap(err)
	}
	rt.Resources = defaultBackendResources()
	if req.BackendResources != nil {
		rt.Resources = *req.BackendResources
	}
	app, err := s.provision(ctx, db, req, rt, provisionSpec{
		key: backendAppKey, name: backendAppName, note: backendAppNote,
		engine: string(base.LoggingBackendTypeVictoriaLogs),
		// The API's network: that is the one it queries on. Never the routing
		// network - VictoriaLogs has no authentication and no tenancy.
		networks: []string{base.NetworkHivepaasLocal},
	}, resp)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	*backend = app
	return ingest, nil
}

// applyCollector provisions or reconciles the collector HivePaaS runs.
func (s *service) applyCollector(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
	cfg *entity.LoggingSettings,
	ingest *logging.Endpoint,
	collector **entity.App,
	resp *loggingservice.SettingApplyResp,
) error {
	spec, err := s.buildCollectSpec(cfg, ingest)
	if err != nil {
		return hperrors.Wrap(err)
	}
	_, collectorImage := releaseImages()
	c, err := logging.NewCollector(logging.CollectorType(cfg.Collector.Type),
		&logging.CollectorConfig{Vlagent: &vlagent.Config{Image: collectorImage}})
	if err != nil {
		return hperrors.Wrap(err)
	}
	rt, err := c.Configure(spec)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if *collector != nil {
		// The credentials first: they are files the new arguments name.
		if err = s.systemAppService.SyncSecrets(ctx, db, *collector, secretFiles(rt)); err != nil {
			return hperrors.Wrap(err)
		}
		return hperrors.Wrap(s.redeploy(ctx, db, req, *collector, rt, resp))
	}

	rt.Resources = logging.Resources{MemoryLimit: logging.DefaultCollectorMemoryLimit.Bytes()}
	app, err := s.provision(ctx, db, req, rt, provisionSpec{
		key: collectorAppKey, name: collectorAppName, note: collectorAppNote,
		engine: string(base.LoggingCollectorTypeVlagent),
		// Only its environment's network. The internal one holds the database,
		// and this runs on every node.
	}, resp)
	if err != nil {
		return hperrors.Wrap(err)
	}
	*collector = app
	return nil
}

// provisionSpec is what tells one app of the stack from the other.
type provisionSpec struct {
	key, name, note, engine string
	// networks are joined besides the environment's own, under the app's key.
	networks []string
}

func (s *service) provision(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
	rt *logging.RuntimeSpec,
	spec provisionSpec,
	resp *loggingservice.SettingApplyResp,
) (*entity.App, error) {
	doc, err := buildAppDoc(rt, spec.engine)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	s.logger.Info("provisioning a logging app", "key", spec.key)
	provisioned, err := s.systemAppService.Provision(ctx, db, &systemappservice.ProvisionReq{
		Env:           loggingEnv,
		Key:           spec.key,
		Name:          spec.name,
		Note:          spec.note,
		Doc:           doc,
		Customize:     customizeSpec(rt, spec.networks, spec.key),
		TriggerUserID: req.TriggerUserID,
	})
	if provisioned != nil && provisioned.Cleanup != nil {
		addCleanup(resp, provisioned.Cleanup)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if provisioned.DeploymentTask != nil {
		resp.Tasks = append(resp.Tasks, provisioned.DeploymentTask)
	}
	resp.Tasks = append(resp.Tasks, provisioned.CertTasks...)
	return provisioned.App, nil
}

// redeploy brings a running app's deployment settings in line with rt, and
// queues the deployment that applies them when anything changed. The image is
// the release's: the updater moves it, and a save puts back one that drifted.
func (s *service) redeploy(
	ctx context.Context,
	db database.IDB,
	req *loggingservice.SettingApplyReq,
	app *entity.App,
	rt *logging.RuntimeSpec,
	resp *loggingservice.SettingApplyResp,
) error {
	command := commandLine(rt.Args)
	task, err := s.systemAppService.Redeploy(ctx, db, &systemappservice.RedeployReq{
		App: app,
		Change: func(settings *entity.AppDeploymentSettings) bool {
			if settings.ImageSource == nil {
				settings.ImageSource = &entity.DeploymentImageSource{}
			}
			if settings.Command == command && settings.ImageSource.Image == rt.Image {
				return false
			}
			settings.Command, settings.ImageSource.Image = command, rt.Image
			return true
		},
		TriggerUserID: req.TriggerUserID,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if task != nil {
		resp.Tasks = append(resp.Tasks, task)
	}
	return nil
}

// secretFiles is rt's credentials as the files the collector's app mounts.
func secretFiles(rt *logging.RuntimeSpec) []*systemappservice.SecretFile {
	files := make([]*systemappservice.SecretFile, 0, len(rt.Secrets))
	for _, secret := range rt.Secrets {
		files = append(files, &systemappservice.SecretFile{
			Key: secret.Key, Path: secret.Path, Value: secret.Value,
		})
	}
	return files
}

// rememberApps writes the apps' ids back into the setting, for the dashboard to
// link to. They are never read to find the apps: an id outlives an app somebody
// deleted in its screen, and the key does not.
func (s *service) rememberApps(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.LoggingSettings,
	backend, collector *entity.App,
) error {
	backendID, collectorID := appID(backend), appID(collector)
	if cfg.BackendAppID == backendID && cfg.CollectorAppID == collectorID {
		return nil
	}
	cfg.BackendAppID, cfg.CollectorAppID = backendID, collectorID
	if err := setting.SetData(cfg); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

func appID(app *entity.App) string {
	if app == nil {
		return ""
	}
	return app.ID
}

// addCleanup chains one more cleanup onto resp's, newest first.
func addCleanup(resp *loggingservice.SettingApplyResp, cleanup func(context.Context) error) {
	previous := resp.Cleanup
	resp.Cleanup = func(ctx context.Context) error {
		err := cleanup(ctx)
		if previous == nil {
			return err
		}
		return errors.Join(err, previous(ctx))
	}
}
