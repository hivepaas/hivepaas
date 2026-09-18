package appprovisionserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/approutingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
)

const (
	routingApplyRetryMax   = 3
	routingApplyRetryDelay = 500 * time.Millisecond
)

func (s *service) ApplyAppConfiguration(
	ctx context.Context,
	db database.IDB,
	req *appprovisionservice.ApplyAppConfigurationReq,
) (*appprovisionservice.ApplyAppConfigurationResp, error) {
	app := req.App
	resp := &appprovisionservice.ApplyAppConfigurationResp{}

	// The environment comes first: the app kind makes shared variables such as
	// HIVEPAAS_PASSWORD, which the app's own variables and its secrets refer to.
	if err := s.applyEnvVars(ctx, db, app); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while applying environment variables")
	}
	configs, secrets, err := s.applySwarmFiles(ctx, db, app)
	resp.Configs, resp.Secrets = configs, secrets
	if err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while applying secrets and config files")
	}
	if err := s.applyRouting(ctx, db, app, req.RefObjects); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while applying routing settings")
	}
	if err := s.applySchedJobs(ctx, db, app); err != nil {
		return resp, hperrors.Wrap(err).WithExtraDetail("while scheduling the app's jobs")
	}
	return resp, nil
}

// applyEnvVars builds and applies the environment of every app in this app's
// scope, not only its own: a variable of another app may refer to this one.
func (s *service) applyEnvVars(ctx context.Context, db database.IDB, app *entity.App) error {
	// In a transaction: no nested transactions, and no concurrency.
	appEnvData, err := s.envVarService.BuildEnvVarsForAllAppsInScope(ctx, db, app.GetObjectScope(),
		false, nil, false, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	errMap := s.envVarService.ApplyEnvVarsForApps(ctx, db, appEnvData, false, false)
	for _, applyErr := range errMap {
		return hperrors.Wrap(applyErr)
	}
	return nil
}

// applySwarmFiles creates the docker configs and secrets the app's settings
// describe, and attaches them to its service.
//
// Only a setting that asks for a file becomes a docker object: a secret without
// one is read through the environment, where a ${secrets.NAME} reference
// resolves it. Creating an object fills in the ids it was created with, and the
// settings are written again so that the app keeps them - without them nothing
// could find the object again to update or remove it.
//
// It returns what was created even when it fails part way, because those objects
// are in docker and outlive the transaction that failed.
func (s *service) applySwarmFiles(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (configRefs []*entity.SwarmConfigRef, secretRefs []*entity.SwarmSecretRef, err error) {
	configSettings := app.GetSettingsByType(base.SettingTypeConfigFile)
	secretSettings := app.GetSettingsByType(base.SettingTypeSecret)
	if len(configSettings) == 0 && len(secretSettings) == 0 {
		return nil, nil, nil
	}

	configFiles := make([]*entity.ConfigFile, 0, len(configSettings))
	for _, setting := range configSettings {
		configFile, parseErr := setting.AsConfigFile()
		if parseErr != nil {
			return nil, nil, hperrors.Wrap(parseErr)
		}
		configFiles = append(configFiles, configFile)
	}
	secrets := make([]*entity.Secret, 0, len(secretSettings))
	for _, setting := range secretSettings {
		secret, parseErr := setting.AsSecret()
		if parseErr != nil {
			return nil, nil, hperrors.Wrap(parseErr)
		}
		secrets = append(secrets, secret)
	}

	configRefs, err = s.clusterSecretService.CreateConfigsForApp(ctx, db, app, configFiles)
	if err != nil {
		return configRefs, nil, hperrors.Wrap(err)
	}
	secretRefs, err = s.clusterSecretService.CreateSecretsForApp(ctx, db, app, secrets)
	if err != nil {
		return configRefs, secretRefs, hperrors.Wrap(err)
	}

	// Writing the settings again is what keeps the ids docker handed back. The
	// entities were parsed from these settings and filled in place, so each one
	// only has to be serialized into the setting it came from.
	persisting := &appservice.PersistingAppData{
		UpsertingSettings: make([]*entity.Setting, 0, len(configSettings)+len(secretSettings)),
	}
	for i, setting := range configSettings {
		if err = setting.SetData(configFiles[i]); err != nil {
			return configRefs, secretRefs, hperrors.Wrap(err)
		}
		persisting.UpsertingSettings = append(persisting.UpsertingSettings, setting)
	}
	for i, setting := range secretSettings {
		if err = setting.SetData(secrets[i]); err != nil {
			return configRefs, secretRefs, hperrors.Wrap(err)
		}
		persisting.UpsertingSettings = append(persisting.UpsertingSettings, setting)
	}
	return configRefs, secretRefs, hperrors.Wrap(s.appService.PersistAppData(ctx, db, persisting))
}

// applyRouting writes the app's routing settings to traefik and to its service.
//
// It retries, which applying routing to an app that has been running does not
// have to: the service was created moments ago and swarm's own allocator is
// still writing to it, so an update carrying the version an inspect has just
// returned comes back as "update out of sequence" - about one create in three on
// a developer machine. Each attempt re-inspects the service and writes the same
// settings, so repeating it changes nothing beyond the version it carries.
func (s *service) applyRouting(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	refObjects *entity.RefObjects,
) error {
	routingSetting := app.GetSettingByType(base.SettingTypeAppRouting)
	if routingSetting == nil {
		return nil
	}
	routingSettings, err := routingSetting.AsAppRoutingSettings()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if refObjects == nil {
		refObjects = entity.NewRefObjects()
	}

	for attempt := range routingApplyRetryMax + 1 {
		if attempt > 0 {
			timer := time.NewTimer(routingApplyRetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return hperrors.Wrap(ctx.Err())
			case <-timer.C:
			}
		}
		_, err = s.appRoutingService.ApplyRoutingSettings(ctx, db, &approutingservice.ApplyAppRoutingReq{
			App:             app,
			RoutingSettings: routingSettings,
			RefObjects:      refObjects,
		})
		if err == nil {
			return nil
		}
	}
	return hperrors.Wrap(err)
}

// applySchedJobs queues the first run of each job the app's settings schedule.
func (s *service) applySchedJobs(ctx context.Context, db database.IDB, app *entity.App) error {
	jobSettings := app.GetSettingsByType(base.SettingTypeSchedJob)
	if len(jobSettings) == 0 {
		return nil
	}
	tx, ok := db.(database.Tx)
	if !ok {
		return hperrors.Wrap(hperrors.ErrInternal).WithMsgLog("db is not transaction")
	}
	return hperrors.Wrap(s.taskQueue.ScheduleTasksForSchedJobs(ctx, tx, jobSettings, false))
}
