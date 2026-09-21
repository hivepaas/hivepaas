package registryserviceimpl

import (
	"context"
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
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
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

	// registryDefaultEnv is the environment a system app lands in: the key
	// project_sync gives an app whose service names none.
	registryDefaultEnv = "default"
)

// planInput is what the plan needs that the setting only points at: the volume's
// name rather than its id, the bucket's credentials rather than its id, and the
// htpasswd file, which is derived from a password nothing else may see.
type planInput struct {
	VolumeName string
	S3         *zotS3Input
	Htpasswd   string
}

// planAppDoc turns the setting into the document the app is built from. It is
// separate from Apply, and pure, because this is where every decision is: what
// gets mounted, what the configuration says, how big the app is allowed to be.
func planAppDoc(cfg *entity.RegistrySettings, in planInput) (appDocInput, error) {
	zotConfig, err := renderZotConfig(cfg, zotConfigInput{S3: in.S3})
	if err != nil {
		return appDocInput{}, hperrors.Wrap(err)
	}

	volumeName := in.VolumeName
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		volumeName = ""
	}

	return appDocInput{
		Name:   registryAppName,
		Key:    base.HivepaasRegistryKey,
		Domain: cfg.Domain,
		// DataSize.String already writes "512mb", which is the spelling every
		// size in a spec document uses.
		MemoryLimit: cfg.MemoryLimit.String(),
		VolumeName:  volumeName,
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
		// Switching it off removes nothing: the images are on a volume or in a
		// bucket, and deleting them is done in the app's own screen.
		return &registryservice.SettingApplyResp{}, nil
	}

	credential, password, err := s.ensureCredential(ctx, db, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	in, err := s.resolveStorage(ctx, db, cfg)
	if err != nil {
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
	if app == nil {
		s.logger.Info("provisioning the system registry", "domain", cfg.Domain,
			"storage", cfg.Storage.Type)
		app, err = s.provision(ctx, db, plan, req.TriggerUserID)
	} else {
		err = s.reconcile(ctx, db, app, plan)
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = s.rememberWhatWasCreated(ctx, db, setting, cfg, app.ID, credential.ID); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &registryservice.SettingApplyResp{App: app}, nil
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
) (*entity.App, error) {
	doc, err := renderAppDoc(plan)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	project, projectEnv, err := s.rootProjectEnv(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()
	resp, err := s.provisionService.ProvisionApp(ctx, db, &appprovisionservice.ProvisionAppReq{
		ProjectID:    project.ID,
		ProjectEnvID: projectEnv.ID,
		AppID:        gofn.Must(ulid.NewStringULID()),
		Name:         plan.Name,
		Status:       base.AppStatusActive,
		Note:         "The registry HivePaaS runs for itself. It is configured in system settings.",
		Configure: func(ctx context.Context, db database.IDB, app *entity.App,
			spec *swarm.ServiceSpec) ([]*entity.Setting, error) {
			built, buildErr := s.specService.BuildApp(ctx, db, &specservice.BuildAppReq{
				App: app, Doc: doc, Spec: spec, TimeNow: timeNow,
			})
			if buildErr != nil {
				return nil, hperrors.Wrap(buildErr)
			}
			return built.Settings, nil
		},
		Deployment: &appprovisionservice.FirstDeployment{
			Source:   base.DeploymentTriggerSourceAPI,
			SourceID: triggerUserID,
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
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
			WithExtraDetail("the registry app has no %s config file", registryConfigFileName)
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
			WithExtraDetail("the registry app has no %s secret", registrySecretName)
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
//
// It looks the app up by key rather than by the stored id, so that an id left
// over from an app somebody deleted does not stop the registry from being created
// again.
func (s *service) loadApp(ctx context.Context, db database.IDB) (*entity.App, error) {
	app, err := s.hpAppService.LoadAppByKey(ctx, db, base.HivepaasRegistryKey,
		bunex.SelectRelation("Settings"),
		bunex.SelectRelation("Project"),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, hperrors.Wrap(err)
	}
	return app, nil
}

// rootProjectEnv is where a system app lives: the hidden hivepaas project, and
// the environment a service that names none lands in.
func (s *service) rootProjectEnv(ctx context.Context, db database.IDB) (
	*entity.Project, *entity.ProjectEnv, error) {
	project, err := s.projectRepo.GetByKey(ctx, db, base.HivepaasProjectKey,
		bunex.SelectRelation("ProjectEnvs"),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	if projectEnv := project.GetEnv(registryDefaultEnv); projectEnv != nil {
		return project, projectEnv, nil
	}
	// An installation whose system apps were synced under another environment
	// name still has to be able to provision this one.
	if len(project.ProjectEnvs) > 0 {
		return project, project.ProjectEnvs[0], nil
	}
	return nil, nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured).
		WithExtraDetail("the %s project has no environment to create the registry in",
			base.HivepaasProjectKey)
}
