package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var switchSource = &entity.App{ID: "src-1", Key: "shop-db", Name: "shop-db"}

func envVarsSetting(id string, vars ...*entity.EnvVar) *entity.Setting {
	return setting(id, base.SettingTypeEnvVar, &entity.EnvVars{Data: vars})
}

// consumerApp is an app of the environment with the settings given.
func consumerApp(id, key string, settings ...*entity.Setting) *entity.App {
	for _, one := range settings {
		one.ObjectID = id
	}
	return &entity.App{ID: id, Key: key, Name: key, Settings: settings}
}

func switchEnvApps() []*entity.App {
	features := setting("features-1", base.SettingTypeAppFeatures, &entity.AppFeatureSettings{
		PreviewSettings: &entity.AppFeaturePreviewSettings{Enabled: true,
			AppsToClone: entity.ObjectIDSlice{{ID: "src-1"}, {ID: "cache-1"}}}})
	deployment := setting("deployment-1", base.SettingTypeAppDeployment, &entity.AppDeploymentSettings{
		ActiveMethod: base.DeploymentMethodImage, RestorePointBeforeDeploy: entity.ObjectID{ID: "src-1"}})
	return []*entity.App{
		switchSource,
		consumerApp("target-1", "shop-db-restored", envVarsSetting("env-t",
			&entity.EnvVar{Key: "UPSTREAM", Value: "${shop-db.HIVEPAAS_HOST}"})),
		consumerApp("web-1", "web", envVarsSetting("env-1",
			&entity.EnvVar{Key: "DATABASE_URL", Value: "${shop-db.HIVEPAAS_DATABASE_URL}"},
			&entity.EnvVar{Key: "LITERAL", Value: "${shop-db.HIVEPAAS_HOST}", IsLiteral: true},
			&entity.EnvVar{Key: "CACHE", Value: "${cache.HIVEPAAS_HOST}"}), features, deployment),
		consumerApp("worker-1", "worker", envVarsSetting("env-2",
			&entity.EnvVar{Key: "DSN", Value: "host=${shop-db.HIVEPAAS_HOST} port=${shop-db.HIVEPAAS_PORT}"})),
		consumerApp("other-1", "other", envVarsSetting("env-3",
			&entity.EnvVar{Key: "DB", Value: "${shop-db-2.HIVEPAAS_HOST}"},
			&entity.EnvVar{Key: "ONLY_LITERAL", Value: "${shop-db.HIVEPAAS_HOST}", IsLiteral: true})),
		{ID: "preview-1", Key: "web-pr-1", ParentID: "web-1", Settings: []*entity.Setting{envVarsSetting("env-p",
			&entity.EnvVar{Key: "DATABASE_URL", Value: "${shop-db.HIVEPAAS_DATABASE_URL}"})}},
	}
}

func consumerIDs(consumers []*switchConsumer) []string {
	out := make([]string, 0, len(consumers))
	for _, consumer := range consumers {
		out = append(out, consumer.app.ID)
	}
	return out
}

// The consumers are the apps that name the source: by a variable that is not
// literal, by a preview's database to clone, or by their deployments' restore
// point. The source, the restored database and previews are not among them.
func TestTheConsumersOfASource(t *testing.T) {
	consumers := findConsumers(switchEnvApps(), switchSource, "target-1")

	assert.Equal(t, []string{"web-1", "worker-1"}, consumerIDs(consumers))
	web := consumers[0]
	assert.Equal(t, []string{"DATABASE_URL"}, web.envKeys)
	assert.True(t, web.previewClone)
	assert.True(t, web.deployRestorePoint)
	worker := consumers[1]
	assert.Equal(t, []string{"DSN"}, worker.envKeys)
	assert.False(t, worker.previewClone)
	assert.False(t, worker.deployRestorePoint)
}

// A switch rewrites only the source's references, repoints the previews and the
// deployments' restore point, and returns the settings it changed.
func TestASwitchRepointsAConsumer(t *testing.T) {
	consumers := findConsumers(switchEnvApps(), switchSource, "target-1")
	target := &entity.App{ID: "target-1", Key: "shop-db-restored"}

	changed := switchConsumerTo(consumers[0], switchSource, target)

	assert.ElementsMatch(t, []string{"env-1", "features-1", "deployment-1"}, settingIDs(changed))
	vars := consumers[0].envSetting.MustAsEnvVars().Data
	assert.Equal(t, "${shop-db-restored.HIVEPAAS_DATABASE_URL}", vars[0].Value)
	assert.Equal(t, "${shop-db.HIVEPAAS_HOST}", vars[1].Value, "a literal stays")
	assert.Equal(t, "${cache.HIVEPAAS_HOST}", vars[2].Value)
	clones := consumers[0].features.MustAsAppFeatureSettings().PreviewSettings.AppsToClone
	assert.Equal(t, []string{"target-1", "cache-1"}, []string{clones[0].ID, clones[1].ID})
	assert.Equal(t, "target-1",
		consumers[0].deployment.MustAsAppDeploymentSettings().RestorePointBeforeDeploy.ID)
	for _, one := range changed {
		assert.Equal(t, 4, one.UpdateVer, "%s is a new version", one.ID)
	}

	changed = switchConsumerTo(consumers[1], switchSource, target)
	assert.Equal(t, []string{"env-2"}, settingIDs(changed))
	assert.Equal(t, "host=${shop-db-restored.HIVEPAAS_HOST} port=${shop-db-restored.HIVEPAAS_PORT}",
		consumers[1].envSetting.MustAsEnvVars().Data[0].Value)
}

// Only a database restored from another, and ready, switches consumers to it.
func TestOnlyAReadyRestoredDatabaseSwitches(t *testing.T) {
	loaded := loadedPostgres(t)
	loaded.app = &entity.App{ID: "target-1", Name: "shop-db-restored"}
	assert.ErrorIs(t, checkSwitchable(loaded), hperrors.ErrAppManagedSwitchNotRestored)

	loaded.managed.Postgres.Restore = &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "src-1"}}
	assert.ErrorIs(t, checkSwitchable(loaded), hperrors.ErrAppManagedSwitchNotReady)

	loaded.managed.Postgres.Restore.ReadyAt = time.Now()
	assert.NoError(t, checkSwitchable(loaded))
}

// The source's jobs, the engine's, are paused; a person's, and those already
// paused, are left.
func TestTheSourcesJobsArePaused(t *testing.T) {
	job := func(id, managedBy string, status base.SettingStatus) *entity.Setting {
		one := setting(id, base.SettingTypeSchedJob, &entity.SchedJob{JobType: base.SchedJobTypeContainerCommand,
			ManagedBy: entity.ObjectID{ID: managedBy}})
		one.Status = status
		return one
	}
	settings := []*entity.Setting{
		job("nightly", "managed-src", base.SettingStatusActive),
		job("health", "managed-src", base.SettingStatusActive),
		job("mine", "", base.SettingStatusActive),
		job("off", "managed-src", base.SettingStatusDisabled),
	}
	now := time.Now().UTC()

	paused := pauseManagedJobs(settings, "managed-src", now)

	assert.Equal(t, []string{"nightly", "health"}, settingIDs(paused))
	for _, one := range paused {
		assert.Equal(t, base.SettingStatusDisabled, one.Status)
		assert.Equal(t, 4, one.UpdateVer)
		assert.Equal(t, now, one.UpdatedAt)
	}
	assert.Equal(t, base.SettingStatusActive, settings[2].Status)
}
