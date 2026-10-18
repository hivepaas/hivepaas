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
			"      swarmRef: {file: {name: "+scriptDir+"/"+script+", mode: 555}}")
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
	assert.Equal(t, secretsDir, env["HIVEPAAS_WALG_SECRETS_DIR"])
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
		assert.Equal(t, secretsDir+"/"+name, file["name"], name)
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
	assert.Equal(t, volumeArchiveDir, env["WALG_FILE_PREFIX"])
	assert.NotContains(t, env, "WALG_S3_PREFIX")
	archive := req.Doc.Deployment.Storage.Mounts[volumeArchiveDir]
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
	assert.Contains(t, info.Detail, scriptDir+"/walg-fetch")
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
			assert.Equal(t, string(base.SchedJobTypeContainerCommand), setting.Kind)
			jobs[setting.ID] = setting.MustAsSchedJob()
		}
	}
	if !assert.NotNil(t, managed) || !assert.Len(t, jobs, 3) {
		return
	}
	data := managed.MustAsAppManagedSettings()
	assert.Equal(t, "postgres", data.Engine)
	assert.Equal(t, 1, data.EngineAPI)
	assert.Equal(t, "storage-1", data.Postgres.Storage.CloudStorage.ID)
	assert.Equal(t, 14, data.Postgres.RetainFull)
	assert.Equal(t, time.Minute, time.Duration(data.Postgres.ArchiveTimeout))
	assert.Equal(t, entity.ManagedRestoreTestContainer, data.Postgres.RestoreTest.Mode)
	assert.NoError(t, data.Decrypt())
	plain, _ := data.Postgres.EncryptionKey.GetPlain()
	assert.Equal(t, "libsodium-key-value", plain)

	nightly := jobs[data.Postgres.Jobs.Nightly.ID]
	health := jobs[data.Postgres.Jobs.Health.ID]
	restoreTest := jobs[data.Postgres.Jobs.RestoreTest.ID]
	for name, job := range map[string]*entity.SchedJob{"nightly": nightly, "health": health, "test": restoreTest} {
		if !assert.NotNil(t, job, name) {
			continue
		}
		assert.Equal(t, managed.ID, job.ManagedBy.ID, name)
		assert.Equal(t, "app-1", job.App.ID, name)
		assert.True(t, job.Notification.FailureUseDefault, "%s notifies when it fails", name)
		assert.False(t, job.Notification.SuccessUseDefault, name)
	}
	assert.Equal(t, "30 1 * * *", nightly.Schedule.CronExpr)
	assert.Equal(t, scriptDir+"/walg-nightly", nightly.Command.Command)
	assert.Equal(t, 5*time.Minute, time.Duration(health.Schedule.Interval))
	assert.Equal(t, scriptDir+"/walg-health", health.Command.Command)
	assert.Equal(t, defaultRestoreTestSchedule, restoreTest.Schedule.CronExpr)
	assert.Equal(t, scriptDir+"/walg-restore-test", restoreTest.Command.Command)
}

func TestConfigureNeedsTheDeploymentSetting(t *testing.T) {
	_, err := New().Configure(configureReq(t))

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
}
