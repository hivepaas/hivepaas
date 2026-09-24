package registryserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

const (
	// registryAppName is what the app is called in the dashboard, and
	// registryConfigFileName and registrySecretName are the settings the
	// configuration and the account live in - the keys the app document gives
	// them, which is how they are found again.
	registryAppName        = "Registry"
	registryConfigFileName = "config.json"
	// This is the name of a setting, not a credential: what it holds is the
	// account file, written at provisioning time.
	registrySecretName = "ZOT_HTPASSWD" //nolint:gosec
)

// planInput is what the plan needs that the setting only points at: the volume's
// name rather than its id, the bucket's credentials rather than its id, and the
// htpasswd file, which is derived from a password nothing else may see.
type planInput struct {
	VolumeID string
	S3       *zotS3Input
	Htpasswd string
	// EnvKeys is every environment key in the installation, which retention needs
	// to count each environment's builds separately.
	EnvKeys []string
}

// planAppDoc turns the setting into the document the app is built from. It is
// separate from Apply, and pure, because this is where every decision is: what
// gets mounted, what the configuration says, how big the app is allowed to be.
func planAppDoc(cfg *entity.RegistrySettings, in planInput) (appDocInput, error) {
	zotConfig, err := renderZotConfig(cfg, zotConfigInput{S3: in.S3, EnvKeys: in.EnvKeys})
	if err != nil {
		return appDocInput{}, hperrors.Wrap(err)
	}

	volumeID := in.VolumeID
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		volumeID = ""
	}

	return appDocInput{
		Name:   registryAppName,
		Key:    base.HivepaasRegistryKey,
		Domain: cfg.Domain,
		// DataSize.String already writes "512mb", which is the spelling every
		// size in a spec document uses.
		MemoryLimit: cfg.MemoryLimit.String(),
		OomScoreAdj: base.OomScoreAdjSystemAddon,
		VolumeID:    volumeID,
		ZotConfig:   string(zotConfig),
		Htpasswd:    in.Htpasswd,
	}, nil
}

// Apply makes the cluster match the stored configuration.
//
// It runs after every save and does nothing that is already done, so a save that
// fails half way leaves work the next save retries rather than a broken registry.
func (s *service) Apply(
	ctx context.Context,
	db database.IDB,
	req *registryservice.SettingApplyReq,
) (_ *registryservice.SettingApplyResp, err error) {
	setting := req.Setting
	if setting == nil {
		setting, err = s.loadSetting(ctx, db)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if setting == nil {
			// Never configured, which is the default: there is nothing to run.
			return &registryservice.SettingApplyResp{}, nil
		}
	}

	cfg, err := setting.AsRegistrySettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !cfg.Enabled {
		return s.teardown(ctx, db, setting, cfg, req)
	}

	credential, password, err := s.ensureCredential(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	in, err := s.resolveStorage(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Read on every save rather than kept anywhere: an environment created since
	// the last save gets its own retention rule from here, and one created after
	// this save falls to the catch-all rule until the next.
	if in.EnvKeys, err = s.projectEnvRepo.ListDistinctKeys(ctx, db); err != nil {
		return nil, hperrors.Wrap(err)
	}

	app, err := s.loadApp(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The account file is built from the one the app is running with: bcrypt is
	// salted, so hashing the same password again would change the file on every
	// save and restart the registry for nothing.
	existing := ""
	if app != nil {
		if existing, err = s.currentHtpasswd(app); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	inGrace := !cfg.CredentialGraceEnds.IsZero() && timeutil.NowUTC().Before(cfg.CredentialGraceEnds)
	if in.Htpasswd, err = htpasswdFor(existing, password, inGrace); err != nil {
		return nil, hperrors.Wrap(err)
	}

	plan, err := planAppDoc(cfg, in)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp := &registryservice.SettingApplyResp{}
	if app == nil {
		s.logger.Info("provisioning the system registry", "domain", cfg.Domain,
			"storage", cfg.Storage.Type)
		app, err = s.provision(ctx, db, plan, req.TriggerUserID, resp)
	} else {
		err = s.reconcile(ctx, db, app, plan)
	}
	// From here on the response goes back with the error too: it carries the
	// cleanup of whatever provisioning already created in docker.
	if err != nil {
		return resp, hperrors.Wrap(err)
	}

	if err = s.rememberWhatWasCreated(ctx, db, setting, cfg, app.ID, credential.ID); err != nil {
		return resp, hperrors.Wrap(err)
	}
	resp.App = app
	return resp, nil
}

// teardown is what switching the registry off does.
//
// It used to do nothing at all, on the grounds that the images are on a volume or
// in a bucket and the app is removed from its own screen. That screen cannot be
// reached: the project the registry lives in is left out of every listing, so an
// operator who switched the registry off was left with an app nothing could
// remove. Taking it down is therefore part of this save - but only when the
// caller says so, because a save is not allowed to tear something down unasked.
func (s *service) teardown(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.RegistrySettings,
	req *registryservice.SettingApplyReq,
) (*registryservice.SettingApplyResp, error) {
	app, err := s.loadApp(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if app == nil {
		// Nothing to take down. The ids may still name an app somebody removed
		// another way, and an id that resolves to nothing is worse than none.
		if err = s.rememberWhatWasCreated(ctx, db, setting, cfg, "", ""); err != nil {
			return nil, hperrors.Wrap(err)
		}
		return &registryservice.SettingApplyResp{}, nil
	}

	if !req.RemoveApp {
		return nil, hperrors.Wrap(hperrors.ErrRegistryAppStillRunning).WithExtraDetail(
			"Switching the registry off removes the app that serves it. Confirm the removal, " +
				"and say whether the images should go with it.")
	}

	credentialID := cfg.RegistryAuthID
	s.logger.Info("removing the system registry",
		"app", app.ID, "removeStorage", req.RemoveStorage)
	if err = s.systemAppService.Remove(ctx, db, app, req.RemoveStorage); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = s.rememberWhatWasCreated(ctx, db, setting, cfg, "", ""); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The credential is deleted by the caller, after the commit: it goes through
	// the settings usecase, which is what refuses to delete one an app still
	// names and what records the deletion.
	return &registryservice.SettingApplyResp{
		RemovedApp:          true,
		RemovedCredentialID: credentialID,
	}, nil
}

// rememberWhatWasCreated writes the ids back into the setting. They are how the
// next Apply finds the credential again, and how the dashboard links to both.
func (s *service) rememberWhatWasCreated(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.RegistrySettings,
	appID, credentialID string,
) error {
	if cfg.AppID == appID && cfg.RegistryAuthID == credentialID {
		return nil
	}

	cfg.AppID, cfg.RegistryAuthID = appID, credentialID
	if err := setting.SetData(cfg); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

// provision creates the app the first time, in the hidden hivepaas project, the
// same way the template usecase creates one: the document becomes the app's
// settings, and the settings become the service.
func (s *service) provision(
	ctx context.Context,
	db database.IDB,
	plan appDocInput,
	triggerUserID string,
	out *registryservice.SettingApplyResp,
) (*entity.App, error) {
	doc, err := renderAppDoc(plan)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.systemAppService.Provision(ctx, db, &systemappservice.ProvisionReq{
		Env:           systemappservice.DefaultEnv,
		Key:           plan.Key,
		Name:          plan.Name,
		Note:          "The registry HivePaaS runs for itself. It is configured in system settings.",
		Doc:           doc,
		TriggerUserID: triggerUserID,
	})
	if resp != nil {
		out.Cleanup = resp.Cleanup
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The first deployment is what replaces the placeholder image with zot's, and
	// the certificate tasks are what the domain needs. Both are scheduled by the
	// caller once this transaction has committed.
	out.DeploymentTask, out.CertTasks = resp.DeploymentTask, resp.CertTasks
	return resp.App, nil
}

// reconcile brings a registry that already exists in line with a saved change.
//
// Only the configuration file and the account can change here. The storage is
// refused by validateSettings, and the domain belongs to the app's routing
// screen, which is the only place that verifies a domain and obtains its
// certificate; what this does about a domain is keep the credential's address in
// step with it.
//
// Both files are swarm objects, which are immutable, so each change is a new
// object and a service update: the registry restarts for a few seconds.
func (s *service) reconcile(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	plan appDocInput,
) error {
	if err := s.applyConfigFile(ctx, db, app, plan.ZotConfig); err != nil {
		return hperrors.Wrap(err)
	}
	return s.applyHtpasswd(ctx, db, app, plan.Htpasswd)
}

// applyConfigFile writes a changed configuration into the app and into the swarm.
// It does nothing when the file is the one already there, which is what makes
// saving the same settings twice cost nothing.
func (s *service) applyConfigFile(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	content string,
) error {
	setting := settingNamed(app.GetSettingsByType(base.SettingTypeConfigFile), registryConfigFileName)
	if setting == nil {
		return hperrors.Wrap(hperrors.ErrRegistryNotConfigured).
			WithExtraDetail("The registry app has no %s config file.", registryConfigFileName)
	}

	current, err := setting.AsConfigFile()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if current.Content == content {
		return nil
	}

	updated := *current
	updated.Content = content
	if err = s.clusterSecretService.UpdateConfigForApp(ctx, db, app, current, &updated); err != nil {
		return hperrors.Wrap(err)
	}
	return s.persistSetting(ctx, db, setting, &updated)
}

// applyHtpasswd does for the account what applyConfigFile does for the
// configuration.
func (s *service) applyHtpasswd(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	content string,
) error {
	setting := settingNamed(app.GetSettingsByType(base.SettingTypeSecret), registrySecretName)
	if setting == nil {
		return hperrors.Wrap(hperrors.ErrRegistryNotConfigured).
			WithExtraDetail("The registry app has no %s secret.", registrySecretName)
	}

	current, err := setting.AsSecret()
	if err != nil {
		return hperrors.Wrap(err)
	}
	plain, err := current.Value.GetPlain()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if plain == content {
		return nil
	}

	updated := *current
	updated.Value = entity.NewEncryptedField(content)
	if err = s.clusterSecretService.UpdateSecretForApp(ctx, db, app, current, &updated); err != nil {
		return hperrors.Wrap(err)
	}
	return s.persistSetting(ctx, db, setting, &updated)
}

func (s *service) persistSetting(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	data entity.SettingData,
) error {
	if err := setting.SetData(data); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

// settingNamed picks one setting out of a type's collection by the key the app
// document gave it.
func settingNamed(settings []*entity.Setting, name string) *entity.Setting {
	for _, setting := range settings {
		if setting.Name == name {
			return setting
		}
	}
	return nil
}

// loadSetting returns the stored configuration, or nil when nothing was saved.
func (s *service) loadSetting(ctx context.Context, db database.IDB) (*entity.Setting, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(),
		base.SettingTypeRegistry, false)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

// loadApp returns the registry's app, or nil when it was never provisioned.
func (s *service) loadApp(ctx context.Context, db database.IDB) (*entity.App, error) {
	app, err := s.systemAppService.LoadApp(ctx, db, base.HivepaasRegistryKey)
	return app, hperrors.Wrap(err)
}
