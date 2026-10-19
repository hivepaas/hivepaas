package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// docYAML is a managed Postgres app as its template renders: its data volume, a
// memory limit, its own variables, and the scripts the engine's API 1 names.
const docYAML = `
deployment:
  source: {activeMethod: image, imageSource: {image: "postgres:17.6-trixie"}}
  storage:
    mounts:
      /var/lib/postgresql: {type: volume, source: vol-data}
  resources:
    limits: {memory: 1gb}
settings:
  envVars:
    data:
      - {k: POSTGRES_DB, v: "${HIVEPAAS_DATABASE_NAME}"}
  configFiles:
SCRIPTS
`

func scriptsYAML(skip string) string {
	var lines []string
	for _, script := range apiScripts {
		if script == skip {
			continue
		}
		lines = append(lines, "    "+script+":\n      content: \"#!/bin/sh\\n\"\n"+
			"      swarmRef: {file: {name: "+walgkit.ScriptDir+"/"+script+", mode: 555}}")
	}
	return strings.Join(lines, "\n")
}

func testDoc(t *testing.T, skipScript string) *specmodel.AppDoc {
	t.Helper()
	doc := &specmodel.AppDoc{}
	assert.NoError(t, yaml.Unmarshal([]byte(strings.Replace(docYAML, "SCRIPTS", scriptsYAML(skipScript), 1)), doc))
	return doc
}

func s3Archive() *managedmodel.Archive {
	return &managedmodel.Archive{S3: &managedmodel.S3Archive{
		Endpoint: "https://s3.example.com", Region: "auto", Bucket: "backups", Prefix: "pg/walg/shop-db",
		AccessKey: "AKIAEXAMPLE", SecretKey: "s3-secret-value", ForcePathStyle: true,
	}}
}

func prepareReq(t *testing.T, archive *managedmodel.Archive) *managedmodel.PrepareReq {
	t.Helper()
	return &managedmodel.PrepareReq{Doc: testDoc(t, ""), Version: "17", Archive: archive,
		EncryptionKey: "libsodium-key-value"}
}

func envOf(doc *specmodel.AppDoc) map[string]string {
	env := map[string]string{}
	block, _ := doc.Settings["envVars"].(map[string]any)
	data, _ := block["data"].([]any)
	for _, entry := range data {
		pair, _ := entry.(map[string]any)
		env[pair["k"].(string)] = pair["v"].(string)
	}
	return env
}

func secretsOf(doc *specmodel.AppDoc) map[string]map[string]any {
	out := map[string]map[string]any{}
	block, _ := doc.Settings["secrets"].(map[string]any)
	for name, entry := range block {
		out[name], _ = entry.(map[string]any)
	}
	return out
}

func TestPrepareDocArchivesToS3(t *testing.T) {
	req := prepareReq(t, s3Archive())

	assert.NoError(t, New().PrepareDoc(req))

	env := envOf(req.Doc)
	assert.Equal(t, "${HIVEPAAS_DATABASE_NAME}", env["POSTGRES_DB"], "the template's own are kept")
	assert.Equal(t, "s3://backups/pg/walg/shop-db", env["WALG_S3_PREFIX"])
	assert.Equal(t, "https://s3.example.com", env["AWS_ENDPOINT"])
	assert.Equal(t, "auto", env["AWS_REGION"])
	assert.Equal(t, "true", env["AWS_S3_FORCE_PATH_STYLE"])
	assert.Equal(t, "true", env["WALG_PREVENT_WAL_OVERWRITE"])
	assert.Equal(t, "zstd", env["WALG_COMPRESSION_METHOD"])
	assert.Equal(t, "hex", env["WALG_LIBSODIUM_KEY_TRANSFORM"])
	assert.Equal(t, walgkit.SecretsDir, env["HIVEPAAS_WALG_SECRETS_DIR"])
	for _, value := range env {
		assert.NotContains(t, value, "s3-secret-value", "no key in the environment")
		assert.NotContains(t, value, "libsodium-key-value")
	}

	secrets := secretsOf(req.Doc)
	assert.Len(t, secrets, 3)
	for name, value := range map[string]string{
		"WALG_AWS_ACCESS_KEY_ID": "AKIAEXAMPLE", "WALG_AWS_SECRET_ACCESS_KEY": "s3-secret-value",
		"WALG_LIBSODIUM_KEY": "libsodium-key-value",
	} {
		assert.Equal(t, value, secrets[name]["value"], name)
		file := secrets[name]["swarmRef"].(map[string]any)["file"].(map[string]any)
		assert.Equal(t, walgkit.SecretsDir+"/"+name, file["name"], name)
		assert.Equal(t, "400", file["mode"], "only the postgres user reads it")
		assert.Equal(t, postgresUID, file["uid"])
	}
	assert.NoError(t, specmodel.CheckBuildable(req.Doc), "what it adds is what a template may say")
}

func TestPrepareDocArchivesToAVolume(t *testing.T) {
	req := prepareReq(t, &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{
		VolumeID: "vol-repo", Subpath: "backups/walg/shop-db"}})

	assert.NoError(t, New().PrepareDoc(req))

	env := envOf(req.Doc)
	assert.Equal(t, walgkit.VolumeArchiveDir, env["WALG_FILE_PREFIX"])
	assert.NotContains(t, env, "WALG_S3_PREFIX")
	archive := req.Doc.Deployment.Storage.Mounts[walgkit.VolumeArchiveDir]
	assert.Equal(t, "vol-repo", archive.Source)
	assert.Equal(t, "backups/walg/shop-db", archive.VolumeOptions.Subpath)
	assert.Contains(t, req.Doc.Deployment.Storage.Mounts, "/var/lib/postgresql", "the data volume stays")
	assert.Equal(t, []string{"WALG_LIBSODIUM_KEY"}, keys(secretsOf(req.Doc)))
	assert.NoError(t, specmodel.CheckBuildable(req.Doc))
}

func TestPrepareDocRefusesATemplateMissingAScript(t *testing.T) {
	req := prepareReq(t, s3Archive())
	req.Doc = testDoc(t, "walg-fetch")

	err := New().PrepareDoc(req)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	info, _ := hperrors.ParseError(err, "en")
	assert.Contains(t, info.Detail, walgkit.ScriptDir+"/walg-fetch")
}

func TestPrepareDocRefusesNoArchive(t *testing.T) {
	req := prepareReq(t, &managedmodel.Archive{})

	assert.Error(t, New().PrepareDoc(req))
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func configureReq(t *testing.T, settings ...*entity.Setting) *managedmodel.ConfigureReq {
	t.Helper()
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	prepared := prepareReq(t, s3Archive())
	assert.NoError(t, New().PrepareDoc(prepared))
	return &managedmodel.ConfigureReq{
		App:      &entity.App{ID: "app-1", Key: "shop-db"},
		Prepared: prepared,
		Storage:  entity.ManagedStorage{CloudStorage: entity.ObjectID{ID: "storage-1"}, Prefix: "pg/walg/shop-db"},
		Settings: settings,
		Now:      time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC),
	}
}

func deploymentSetting() *entity.Setting {
	setting := &entity.Setting{ID: "dep-1", Type: base.SettingTypeAppDeployment,
		Version: entity.CurrentAppDeploymentSettingsVersion}
	setting.MustSetData(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage,
		ImageSource: &entity.DeploymentImageSource{Image: "postgres:17.6-trixie"}})
	return setting
}

func TestConfigureGivesPostgresItsArchiveArguments(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)

	_, err := New().Configure(req)

	assert.NoError(t, err)
	command := deployment.MustAsAppDeploymentSettings().Command
	for _, arg := range []string{"postgres", "-c wal_level=replica", "-c archive_mode=on",
		"-c archive_timeout=60s", `-c "archive_command=/opt/hivepaas/bin/walg-run wal-push %p"`,
		"-c shared_buffers=256MB"} {
		assert.Contains(t, command, arg)
	}
}

func TestConfigureLeavesSharedBuffersToPostgresWithoutALimit(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)
	req.Prepared.Doc.Deployment.Resources = nil

	_, err := New().Configure(req)

	assert.NoError(t, err)
	assert.NotContains(t, deployment.MustAsAppDeploymentSettings().Command, "shared_buffers")
}

func TestConfigureMakesTheManagedSettingAndItsJobs(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{BaseBackupSchedule: "30 1 * * *", RetainFull: 14}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	var managed *entity.Setting
	jobs := map[string]*entity.SchedJob{}
	for _, setting := range resp.Settings {
		assert.Equal(t, "app-1", setting.ObjectID)
		assert.Equal(t, base.ObjectScopeApp, setting.Scope)
		switch setting.Type { //nolint:exhaustive // the engine makes two types
		case base.SettingTypeAppManaged:
			managed = setting
		case base.SettingTypeSchedJob:
			jobs[setting.ID] = setting.MustAsSchedJob()
			assert.Equal(t, string(jobs[setting.ID].JobType), setting.Kind)
		}
	}
	if !assert.NotNil(t, managed) || !assert.Len(t, jobs, 3) {
		return
	}
	data := managed.MustAsAppManagedSettings()
	assert.Equal(t, "postgres", data.Engine)
	assert.Equal(t, 1, data.EngineAPI)
	assert.Equal(t, "storage-1", data.Database.Storage.CloudStorage.ID)
	assert.Equal(t, 14, data.Database.RetainFull)
	assert.Equal(t, time.Minute, time.Duration(data.Database.ArchiveTimeout))
	assert.Equal(t, entity.ManagedRestoreTestContainer, data.Database.RestoreTest.Mode)
	assert.NoError(t, data.Decrypt())
	plain, _ := data.Database.EncryptionKey.GetPlain()
	assert.Equal(t, "libsodium-key-value", plain)

	nightly := jobs[data.Database.Jobs.Nightly.ID]
	health := jobs[data.Database.Jobs.Health.ID]
	restoreTest := jobs[data.Database.Jobs.RestoreTest.ID]
	for name, job := range map[string]*entity.SchedJob{"nightly": nightly, "health": health, "test": restoreTest} {
		if !assert.NotNil(t, job, name) {
			continue
		}
		assert.Equal(t, managed.ID, job.ManagedBy.ID, name)
		assert.Equal(t, "app-1", job.App.ID, name)
		assert.True(t, job.Notification.FailureUseDefault, "%s notifies when it fails", name)
		assert.False(t, job.Notification.SuccessUseDefault, name)
	}
	assert.Equal(t, base.SchedJobTypeManagedNightly, nightly.JobType, "its retention is the engine's")
	assert.Equal(t, base.SchedJobTypeContainerCommand, health.JobType)
	assert.Equal(t, base.SchedJobTypeContainerCommand, restoreTest.JobType)
	assert.Equal(t, "30 1 * * *", nightly.Schedule.CronExpr)
	assert.Equal(t, walgkit.ScriptDir+"/walg-nightly", nightly.Command.Command)
	assert.Equal(t, 5*time.Minute, time.Duration(health.Schedule.Interval))
	assert.Equal(t, walgkit.ScriptDir+"/walg-health", health.Command.Command)
	assert.Equal(t, walgkit.DefaultRestoreTestSchedule, restoreTest.Schedule.CronExpr)
	assert.Equal(t, walgkit.ScriptDir+"/walg-restore-test", restoreTest.Command.Command)
}

func TestConfigureNeedsTheDeploymentSetting(t *testing.T) {
	_, err := New().Configure(configureReq(t))

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}

// What the engine owns is recorded as it is made, so that the settings screens
// can refuse to change it without knowing anything of Postgres.
func TestConfigureRecordsWhatTheEngineOwns(t *testing.T) {
	deployment := deploymentSetting()
	secret := &entity.Setting{ID: "secret-1", Type: base.SettingTypeSecret}
	script := &entity.Setting{ID: "config-1", Type: base.SettingTypeConfigFile}
	scriptMount := &entity.Setting{ID: "mount-1", Type: base.SettingTypeAppSettingMount}
	routing := &entity.Setting{ID: "routing-1", Type: base.SettingTypeAppRouting}
	req := configureReq(t, deployment, secret, script, scriptMount, routing)

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	locks := resp.Settings[0].MustAsAppManagedSettings().Locks
	assert.Equal(t, "postgres:17.6-trixie", locks.Image)
	assert.Equal(t, "17", locks.ImageLine)
	assert.Equal(t, deployment.MustAsAppDeploymentSettings().Command, locks.Command)
	assert.Equal(t, []string{"/var/lib/postgresql"}, locks.Mounts)
	assert.ElementsMatch(t, []string{"secret-1", "config-1", "mount-1"}, locks.Settings, "not the routing")
	for _, key := range []string{"WALG_S3_PREFIX", "AWS_ENDPOINT", "WALG_LIBSODIUM_KEY_TRANSFORM", "PGDATA"} {
		assert.Contains(t, locks.EnvKeys, key)
	}
	assert.NotContains(t, locks.EnvKeys, "POSTGRES_DB", "the template's own stay the person's")
}

func jobEnv(job *entity.SchedJob, key string) string {
	for _, env := range job.Command.EnvVars {
		if env.Key == key {
			return env.Value
		}
	}
	return ""
}

// What the scripts are told comes in their job's environment, so changing it is
// changing the job.
func TestConfigureTellsTheScriptsTheirOptions(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{RetainFull: 14, RestoreTestQuery: "SELECT count(*) FROM orders"}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	data := resp.Settings[0].MustAsAppManagedSettings()
	jobs := map[string]*entity.SchedJob{}
	for _, setting := range resp.Settings[1:] {
		jobs[setting.ID] = setting.MustAsSchedJob()
	}
	assert.Equal(t, "14", jobEnv(jobs[data.Database.Jobs.Nightly.ID], walgkit.EnvRetainFull))
	assert.Equal(t, "SELECT count(*) FROM orders",
		jobEnv(jobs[data.Database.Jobs.RestoreTest.ID], walgkit.EnvRestoreCheckQuery))
	assert.Equal(t, "SELECT count(*) FROM orders", data.Database.RestoreTest.CheckQuery)
}

func reconfigureSetup(t *testing.T) (*entity.AppManagedSettings, map[string]*entity.Setting) {
	t.Helper()
	resp, err := New().Configure(configureReq(t, deploymentSetting()))
	assert.NoError(t, err)
	jobs := map[string]*entity.Setting{}
	for _, setting := range resp.Settings[1:] {
		jobs[setting.ID] = setting
	}
	return resp.Settings[0].MustAsAppManagedSettings(), jobs
}

func TestReconfigureChangesTheJobsItMade(t *testing.T) {
	managed, jobs := reconfigureSetup(t)

	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{RetainFull: new(30), BaseBackupSchedule: new("0 1 * * *"),
			RestoreTestSchedule: new("0 5 * * 6"), RestoreTestQuery: new("SELECT 1")}})

	assert.NoError(t, err)
	assert.Equal(t, 30, managed.Database.RetainFull)
	assert.Equal(t, "SELECT 1", managed.Database.RestoreTest.CheckQuery)
	nightly := jobs[managed.Database.Jobs.Nightly.ID].MustAsSchedJob()
	test := jobs[managed.Database.Jobs.RestoreTest.ID].MustAsSchedJob()
	assert.Equal(t, "0 1 * * *", nightly.Schedule.CronExpr)
	assert.Equal(t, "30", jobEnv(nightly, walgkit.EnvRetainFull))
	assert.Equal(t, "0 5 * * 6", test.Schedule.CronExpr)
	assert.Equal(t, "SELECT 1", jobEnv(test, walgkit.EnvRestoreCheckQuery))
	assert.ElementsMatch(t, []string{managed.Database.Jobs.Nightly.ID, managed.Database.Jobs.RestoreTest.ID},
		ids(resp.Changed), "the health check is left alone")
	assert.ElementsMatch(t, ids(resp.Changed), ids(resp.Rescheduled), "both schedules moved")
}

// What a request leaves out stays as it was.
func TestReconfigureLeavesWhatItIsNotAsked(t *testing.T) {
	managed, jobs := reconfigureSetup(t)

	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{RetainFull: new(3)}})

	assert.NoError(t, err)
	nightly := jobs[managed.Database.Jobs.Nightly.ID].MustAsSchedJob()
	assert.Equal(t, walgkit.DefaultBaseBackupSchedule, nightly.Schedule.CronExpr)
	assert.Equal(t, "3", jobEnv(nightly, walgkit.EnvRetainFull))
	assert.Len(t, resp.Changed, 1)
	assert.Empty(t, resp.Rescheduled, "no schedule moved")
}

func ids(settings []*entity.Setting) []string {
	out := make([]string, 0, len(settings))
	for _, s := range settings {
		out = append(out, s.ID)
	}
	return out
}

func TestReconfigureClearsTheCheckQuery(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{RestoreTestQuery: "SELECT 1"}
	configured, err := New().Configure(req)
	assert.NoError(t, err)
	managed := configured.Settings[0].MustAsAppManagedSettings()
	jobs := map[string]*entity.Setting{}
	for _, setting := range configured.Settings[1:] {
		jobs[setting.ID] = setting
	}

	_, err = New().Reconfigure(&managedmodel.ReconfigureReq{Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{RestoreTestQuery: new("")}})

	assert.NoError(t, err)
	assert.Empty(t, managed.Database.RestoreTest.CheckQuery)
	assert.Empty(t, jobEnv(jobs[managed.Database.Jobs.RestoreTest.ID].MustAsSchedJob(), walgkit.EnvRestoreCheckQuery))
}

func TestReconfigureRefusesABadSchedule(t *testing.T) {
	managed, jobs := reconfigureSetup(t)

	_, err := New().Reconfigure(&managedmodel.ReconfigureReq{Managed: managed, Jobs: jobs,
		Change: &managedmodel.SettingsChange{BaseBackupSchedule: new("at two")}})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
}

func TestArchiveSecretsAreTheS3Keys(t *testing.T) {
	assert.Equal(t, map[string]string{walgkit.SecretAccessKey: "AKIAEXAMPLE", walgkit.SecretSecretKey: "s3-secret-value"},
		New().ArchiveSecrets(s3Archive()))
	assert.Empty(t, New().ArchiveSecrets(&managedmodel.Archive{Volume: &managedmodel.VolumeArchive{}}),
		"a volume archive has none")
}
