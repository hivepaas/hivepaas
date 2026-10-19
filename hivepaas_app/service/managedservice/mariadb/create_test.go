package mariadb

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

// docYAML is a managed MariaDB app as its template renders: its data volume, a
// memory limit, its own variables, and the scripts the engine's API 1 names.
const docYAML = `
deployment:
  source: {activeMethod: image, imageSource: {image: "mariadb:11.8.9-noble"}}
  storage:
    mounts:
      /var/lib/mysql: {type: volume, source: vol-data}
  resources:
    limits: {memory: 1gb}
settings:
  envVars:
    data:
      - {k: MARIADB_DATABASE, v: "${HIVEPAAS_DATABASE_NAME}"}
      - {k: MARIADB_MYSQL_LOCALHOST_USER, v: "1"}
      - {k: MARIADB_MYSQL_LOCALHOST_GRANTS, v: ALL}
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
		Endpoint: "https://s3.example.com", Region: "auto", Bucket: "backups", Prefix: "walg/shop-db/app-1",
		AccessKey: "AKIAEXAMPLE", SecretKey: "s3-secret-value", ForcePathStyle: true,
	}}
}

func prepareReq(t *testing.T, archive *managedmodel.Archive) *managedmodel.PrepareReq {
	t.Helper()
	return &managedmodel.PrepareReq{Doc: testDoc(t, ""), Version: "11.8", Archive: archive,
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

func envCount(doc *specmodel.AppDoc, key string) int {
	n := 0
	block, _ := doc.Settings["envVars"].(map[string]any)
	data, _ := block["data"].([]any)
	for _, entry := range data {
		if pair, _ := entry.(map[string]any); pair["k"] == key {
			n++
		}
	}
	return n
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
	assert.Equal(t, "${HIVEPAAS_DATABASE_NAME}", env["MARIADB_DATABASE"], "the template's own are kept")
	assert.Equal(t, "s3://backups/walg/shop-db/app-1", env["WALG_S3_PREFIX"])
	assert.Equal(t, "https://s3.example.com", env["AWS_ENDPOINT"])
	assert.Equal(t, "auto", env["AWS_REGION"])
	assert.Equal(t, "true", env["AWS_S3_FORCE_PATH_STYLE"])
	assert.Equal(t, "zstd", env["WALG_COMPRESSION_METHOD"])
	assert.Equal(t, "hex", env["WALG_LIBSODIUM_KEY_TRANSFORM"])
	assert.Equal(t, walgkit.SecretsDir, env["HIVEPAAS_WALG_SECRETS_DIR"])
	assert.Equal(t, "60", env["HIVEPAAS_ARCHIVE_TIMEOUT"], "the archive loop's period, in seconds")
	assert.NotContains(t, env, "WALG_PREVENT_WAL_OVERWRITE", "Postgres's: binlog-push has none")
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
		assert.Equal(t, "400", file["mode"], "only the mysql user reads it")
		assert.Equal(t, mysqlUID, file["uid"], "the archive loop runs as mysql")
	}
	assert.NoError(t, specmodel.CheckBuildable(req.Doc), "what it adds is what a template may say")
}

func TestPrepareDocPassesTheArchiveTimeout(t *testing.T) {
	req := prepareReq(t, s3Archive())
	req.Options = &managedmodel.Options{ArchiveTimeout: 2 * time.Minute}

	assert.NoError(t, New().PrepareDoc(req))

	assert.Equal(t, "120", envOf(req.Doc)["HIVEPAAS_ARCHIVE_TIMEOUT"])
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
	assert.Contains(t, req.Doc.Deployment.Storage.Mounts, "/var/lib/mysql", "the data volume stays")
	assert.Len(t, secretsOf(req.Doc), 1, "the key alone")
	assert.NoError(t, specmodel.CheckBuildable(req.Doc))
}

// The scripts reach the server as mysql@localhost, which the image makes from
// two variables of the template's: a template without them gets them.
func TestPrepareDocMakesSureOfTheScriptsAccount(t *testing.T) {
	req := prepareReq(t, s3Archive())
	block := req.Doc.Settings["envVars"].(map[string]any)
	block["data"] = block["data"].([]any)[:1]

	assert.NoError(t, New().PrepareDoc(req))

	env := envOf(req.Doc)
	assert.Equal(t, "1", env["MARIADB_MYSQL_LOCALHOST_USER"])
	assert.Equal(t, "ALL", env["MARIADB_MYSQL_LOCALHOST_GRANTS"])

	kept := prepareReq(t, s3Archive())
	assert.NoError(t, New().PrepareDoc(kept))
	assert.Equal(t, 1, envCount(kept.Doc, "MARIADB_MYSQL_LOCALHOST_USER"), "the template's, not twice")
}

func TestPrepareDocRefusesATemplateMissingAScript(t *testing.T) {
	req := prepareReq(t, s3Archive())
	req.Doc = testDoc(t, "walg-archive")

	err := New().PrepareDoc(req)

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateInvalid)
	info, _ := hperrors.ParseError(err, "en")
	assert.Contains(t, info.Detail, walgkit.ScriptDir+"/walg-archive")
}

func TestPrepareDocRefusesNoArchive(t *testing.T) {
	req := prepareReq(t, &managedmodel.Archive{})

	assert.Error(t, New().PrepareDoc(req))
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
		Storage:  entity.ManagedStorage{CloudStorage: entity.ObjectID{ID: "storage-1"}, Prefix: "walg/shop-db/app-1"},
		Settings: settings,
		Now:      time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC),
	}
}

func deploymentSetting() *entity.Setting {
	setting := &entity.Setting{ID: "dep-1", Type: base.SettingTypeAppDeployment,
		Version: entity.CurrentAppDeploymentSettingsVersion}
	setting.MustSetData(&entity.AppDeploymentSettings{ActiveMethod: base.DeploymentMethodImage,
		ImageSource: &entity.DeploymentImageSource{Image: "mariadb:11.8.9-noble"}})
	return setting
}

func TestConfigureGivesMariaDBItsBinaryLogArguments(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)

	_, err := New().Configure(req)

	assert.NoError(t, err)
	source := deployment.MustAsAppDeploymentSettings()
	assert.Equal(t, "mariadbd --datadir=/var/lib/mysql/data --log-bin=mariadb-bin --binlog-format=ROW "+
		"--server-id=1 --log-slave-updates --binlog-expire-logs-seconds=172800 --innodb-buffer-pool-size=256M",
		source.Command)
	assert.Equal(t, walgkit.ScriptDir+"/walg-entrypoint", source.Entrypoint)
}

func TestConfigureLeavesTheBufferPoolToMariaDBWithoutALimit(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)
	req.Prepared.Doc.Deployment.Resources = nil

	_, err := New().Configure(req)

	assert.NoError(t, err)
	assert.NotContains(t, deployment.MustAsAppDeploymentSettings().Command, "buffer-pool")
}

func TestConfigureMakesTheManagedSettingAndItsJobs(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{BaseBackupSchedule: "30 1 * * *", RetainFull: 14,
		RestoreTestQuery: "SELECT COUNT(*) FROM orders"}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	if !assert.Len(t, resp.Settings, 4) {
		return
	}
	data := resp.Settings[0].MustAsAppManagedSettings()
	assert.Equal(t, "mariadb", data.Engine)
	assert.Equal(t, 1, data.EngineAPI)
	assert.Equal(t, "storage-1", data.Database.Storage.CloudStorage.ID)
	assert.Equal(t, 14, data.Database.RetainFull)
	assert.Equal(t, time.Minute, time.Duration(data.Database.ArchiveTimeout))
	jobs := map[string]*entity.SchedJob{}
	for _, setting := range resp.Settings[1:] {
		jobs[setting.ID] = setting.MustAsSchedJob()
	}
	nightly := jobs[data.Database.Jobs.Nightly.ID]
	health := jobs[data.Database.Jobs.Health.ID]
	restoreTest := jobs[data.Database.Jobs.RestoreTest.ID]
	assert.Equal(t, base.SchedJobTypeManagedNightly, nightly.JobType)
	assert.Equal(t, "30 1 * * *", nightly.Schedule.CronExpr)
	assert.Equal(t, walgkit.ScriptDir+"/walg-nightly", nightly.Command.Command)
	assert.Equal(t, walgkit.ScriptDir+"/walg-health", health.Command.Command)
	assert.Equal(t, walgkit.ScriptDir+"/walg-restore-test", restoreTest.Command.Command)
	assert.Equal(t, "SELECT COUNT(*) FROM orders", restoreTest.Command.EnvVars[0].Value)
}

// What the engine owns is recorded as it is made: the archive's variables, and
// the account the scripts reach the server as.
func TestConfigureRecordsWhatTheEngineOwns(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	locks := resp.Settings[0].MustAsAppManagedSettings().Locks
	assert.Equal(t, "mariadb:11.8.9-noble", locks.Image)
	assert.Equal(t, "11.8", locks.ImageLine)
	assert.Equal(t, deployment.MustAsAppDeploymentSettings().Command, locks.Command)
	assert.Equal(t, []string{"/var/lib/mysql"}, locks.Mounts)
	for _, key := range []string{"WALG_S3_PREFIX", "HIVEPAAS_ARCHIVE_TIMEOUT", "MARIADB_MYSQL_LOCALHOST_USER",
		"MARIADB_MYSQL_LOCALHOST_GRANTS"} {
		assert.Contains(t, locks.EnvKeys, key)
	}
	assert.NotContains(t, locks.EnvKeys, "MARIADB_DATABASE", "the template's own stay the person's")
}

// The dump is mariadb-dump's, as the scripts' account, into the database's own.
func TestConfigureMakesTheDumpJob(t *testing.T) {
	req := configureReq(t, deploymentSetting())
	req.Prepared.Options = &managedmodel.Options{DumpRepo: "repo-1"}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	data := resp.Settings[0].MustAsAppManagedSettings()
	var dump *entity.SchedJob
	for _, setting := range resp.Settings[1:] {
		if setting.ID == data.Database.Jobs.Dump.ID {
			dump = setting.MustAsSchedJob()
		}
	}
	if !assert.NotNil(t, dump) {
		return
	}
	backup := dump.DataBackup
	assert.Equal(t, "repo-1", backup.TargetRepository.ID)
	assert.Equal(t, "shop-db.sql", backup.SourceFileName)
	assert.Contains(t, backup.SourceCommand.Command, "gosu mysql mariadb-dump --user=mysql --single-transaction")
	assert.Contains(t, backup.SourceCommand.Command, `"$MARIADB_DATABASE"`)
	assert.Contains(t, backup.RestoreCommand.Command, `gosu mysql mariadb --user=mysql "$MARIADB_DATABASE"`)
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
	schedule, retain := "15 3 * * *", 10

	resp, err := New().Reconfigure(&managedmodel.ReconfigureReq{App: &entity.App{ID: "app-1"}, Managed: managed,
		Jobs: jobs, Change: &managedmodel.SettingsChange{BaseBackupSchedule: &schedule, RetainFull: &retain}})

	assert.NoError(t, err)
	assert.Equal(t, 10, managed.Database.RetainFull)
	if assert.Len(t, resp.Rescheduled, 1) {
		assert.Equal(t, "15 3 * * *", resp.Rescheduled[0].MustAsSchedJob().Schedule.CronExpr)
	}
}
