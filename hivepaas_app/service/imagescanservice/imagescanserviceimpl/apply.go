package imagescanserviceimpl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// tokenBytes is how much randomness the server's token carries.
const tokenBytes = 32

func (s *service) Apply(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.SettingApplyReq,
) (_ *imagescanservice.SettingApplyResp, err error) {
	setting := req.Setting
	if setting == nil {
		if setting, _, err = s.Settings(ctx, db); err != nil {
			return nil, err
		}
		if setting == nil {
			// Never configured, which is the default: there is nothing to run.
			return &imagescanservice.SettingApplyResp{}, nil
		}
	}
	cfg, err := setting.AsImageScanSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	server, err := s.systemAppService.LoadApp(ctx, db, base.HivepaasTrivyKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	runner, err := s.systemAppService.LoadApp(ctx, db, base.HivepaasTrivyRunnerKey)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &imagescanservice.SettingApplyResp{}
	if !cfg.Enabled {
		// The runner first: nothing asks a server that is going away.
		for _, app := range []*entity.App{runner, server} {
			if app == nil {
				continue
			}
			s.logger.Info("removing an image scanning app", "app", app.ID)
			if err = s.systemAppService.Remove(ctx, db, app, false); err != nil {
				return nil, hperrors.Wrap(err)
			}
			resp.Removed = true
		}
		return resp, s.rememberApps(ctx, db, setting, cfg, nil, nil)
	}

	token, err := tokenOf(server, runner)
	if err != nil {
		return nil, err
	}
	image := trivyImage()

	server, err = s.applyServer(ctx, db, req, server, appDocInput{
		Image: image, Entrypoint: "sh", Command: serverCommand(cfg.DBRepository), Token: token,
		Resources: systemappservice.Resources{
			CPULimit:    entity.DefaultImageScanServerCPULimit,
			MemoryLimit: entity.DefaultImageScanServerMemoryLimit.Bytes(),
		},
	}, resp)
	// From here on the response goes back with the error too: it carries the
	// cleanup of whatever provisioning already created in docker.
	if err != nil {
		return resp, err
	}
	scanResources := gofn.PtrValueOrEmpty(req.ScanResources)
	if req.ScanResources == nil {
		scanResources = defaultScanResources()
	}
	runner, err = s.applyRunner(ctx, db, req, runner, appDocInput{
		Image: image, Entrypoint: "sh", Command: runnerCommand(), Token: token, Resources: scanResources,
	}, resp)
	if err != nil {
		return resp, err
	}
	return resp, s.rememberApps(ctx, db, setting, cfg, server, runner)
}

func (s *service) applyServer(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.SettingApplyReq,
	server *entity.App,
	in appDocInput,
	resp *imagescanservice.SettingApplyResp,
) (*entity.App, error) {
	if server == nil {
		return s.provision(ctx, db, req, provisionSpec{
			key: base.HivepaasTrivyKey, name: serverAppName, note: serverAppNote, customize: customizeServer,
		}, in, resp)
	}
	return server, s.reconcile(ctx, db, req, server, in, resp, false)
}

func (s *service) applyRunner(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.SettingApplyReq,
	runner *entity.App,
	in appDocInput,
	resp *imagescanservice.SettingApplyResp,
) (*entity.App, error) {
	if runner == nil {
		app, err := s.provision(ctx, db, req, provisionSpec{
			key: base.HivepaasTrivyRunnerKey, name: runnerAppName, note: runnerAppNote, customize: customizeRunner,
		}, in, resp)
		if err != nil {
			return nil, err
		}
		_, err = s.ensureHostAccess(ctx, db, app)
		return app, err
	}
	if req.ScanResources != nil {
		if err := s.systemAppService.SetResources(ctx, runner, *req.ScanResources); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	// Access given again is a socket the service does not have yet: only a
	// deployment attaches it.
	granted, err := s.ensureHostAccess(ctx, db, runner)
	if err != nil {
		return nil, err
	}
	return runner, s.reconcile(ctx, db, req, runner, in, resp, granted)
}

// provisionSpec is what tells one app of image scanning from the other.
type provisionSpec struct {
	key, name, note string
	customize       func(spec *swarm.ServiceSpec) error
}

func (s *service) provision(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.SettingApplyReq,
	spec provisionSpec,
	in appDocInput,
	resp *imagescanservice.SettingApplyResp,
) (*entity.App, error) {
	doc, err := buildAppDoc(in)
	if err != nil {
		return nil, err
	}
	s.logger.Info("provisioning an image scanning app", "key", spec.key)
	provisioned, err := s.systemAppService.Provision(ctx, db, &systemappservice.ProvisionReq{
		Env:           scanEnv,
		Key:           spec.key,
		Name:          spec.name,
		Note:          spec.note,
		Doc:           doc,
		Customize:     spec.customize,
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

// reconcile brings an app that exists in line with what it should run: the
// token's file, and the image and command of its deployment, which a save
// puts back when they drifted. force deploys it though neither changed.
func (s *service) reconcile(
	ctx context.Context,
	db database.IDB,
	req *imagescanservice.SettingApplyReq,
	app *entity.App,
	in appDocInput,
	resp *imagescanservice.SettingApplyResp,
	force bool,
) error {
	err := s.systemAppService.SyncSecrets(ctx, db, app, []*systemappservice.SecretFile{
		{Key: tokenKey, Path: tokenPath, Value: in.Token},
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	task, err := s.systemAppService.Redeploy(ctx, db, &systemappservice.RedeployReq{
		App: app,
		Change: func(settings *entity.AppDeploymentSettings) bool {
			if settings.ImageSource == nil {
				settings.ImageSource = &entity.DeploymentImageSource{}
			}
			if !force && settings.ImageSource.Image == in.Image && settings.Entrypoint == in.Entrypoint &&
				settings.Command == in.Command {
				return false
			}
			settings.ImageSource.Image, settings.Entrypoint, settings.Command = in.Image, in.Entrypoint, in.Command
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

// ensureHostAccess gives the runner the node's Docker socket, as an app's
// Docker API access in host mode: every deployment of an app takes away a
// socket its access does not give. Trivy reads the node's images through it.
// It says whether the access was given now.
func (s *service) ensureHostAccess(ctx context.Context, db database.IDB, app *entity.App) (bool, error) {
	if current := app.GetSettingByType(base.SettingTypeAppDockerAPI); current != nil {
		access, err := current.AsAppDockerAPISettings()
		if err == nil && access.IsHostMode() && current.Status == base.SettingStatusActive {
			return false, nil
		}
		if err = s.settingRepo.DeleteHard(ctx, db, bunex.DeleteWhere("id = ?", current.ID)); err != nil {
			return false, hperrors.Wrap(err)
		}
	}
	timeNow := timeutil.NowUTC()
	setting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  app.ID,
		Type:      base.SettingTypeAppDockerAPI,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentAppDockerAPIVersion,
		UpdateVer: 1,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	if err := setting.SetData(&entity.AppDockerAPISettings{Mode: entity.DockerAPIModeHost}); err != nil {
		return false, hperrors.Wrap(err)
	}
	if err := s.settingRepo.Insert(ctx, db, setting); err != nil {
		return false, hperrors.Wrap(err)
	}
	app.Settings = append(app.Settings, setting)
	return true, nil
}

// tokenOf is the token the apps hold, so a save does not restart them for a
// new one; a new token when neither holds one.
func tokenOf(apps ...*entity.App) (string, error) {
	for _, app := range apps {
		if app == nil {
			continue
		}
		for _, setting := range app.GetSettingsByType(base.SettingTypeSecret) {
			if setting.Name != tokenKey {
				continue
			}
			secret, err := setting.AsSecret()
			if err != nil {
				return "", hperrors.Wrap(err)
			}
			token, err := secret.Value.GetPlain()
			if err != nil {
				return "", hperrors.Wrap(err)
			}
			if token != "" {
				return token, nil
			}
		}
	}
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", hperrors.Wrap(err)
	}
	return hex.EncodeToString(buf), nil
}

// rememberApps writes the apps' ids back into the setting, for the dashboard to
// link to. They are never read to find the apps: the keys are.
func (s *service) rememberApps(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.ImageScanSettings,
	server, runner *entity.App,
) error {
	serverID, runnerID := appID(server), appID(runner)
	if cfg.ServerAppID == serverID && cfg.RunnerAppID == runnerID {
		return nil
	}
	cfg.ServerAppID, cfg.RunnerAppID = serverID, runnerID
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
func addCleanup(resp *imagescanservice.SettingApplyResp, cleanup func(context.Context) error) {
	previous := resp.Cleanup
	resp.Cleanup = func(ctx context.Context) error {
		err := cleanup(ctx)
		if previous == nil {
			return err
		}
		return errors.Join(err, previous(ctx))
	}
}

// defaultScanResources is what a runner gets when nothing says otherwise.
func defaultScanResources() systemappservice.Resources {
	return systemappservice.Resources{
		CPULimit:    entity.DefaultImageScanCPULimit,
		MemoryLimit: entity.DefaultImageScanMemoryLimit.Bytes(),
	}
}
