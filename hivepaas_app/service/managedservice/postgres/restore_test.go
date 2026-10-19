package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

var targetTime = time.Date(2026, 10, 18, 14, 32, 0, 0, time.UTC)

func sourceArchive() *managedmodel.Archive {
	return &managedmodel.Archive{S3: &managedmodel.S3Archive{
		Endpoint: "https://s3.example.com", Region: "auto", Bucket: "backups", Prefix: "pg/walg/shop-db",
		AccessKey: "AKIA-SOURCE", SecretKey: "source-secret-value", ForcePathStyle: true,
	}}
}

// restoreReq is a restore of shop-db into a new app, which archives to its own
// prefix of the same bucket.
func restoreReq(t *testing.T, target entity.ManagedRestoreTarget) *managedmodel.PrepareReq {
	t.Helper()
	own := s3Archive()
	own.S3.Prefix = "pg/walg/shop-db-restored"
	req := prepareReq(t, own)
	req.Restore = &managedmodel.RestoreSource{SourceAppID: "app-1", Archive: sourceArchive(),
		EncryptionKey: "source-libsodium-key", Target: target, PointAt: targetTime.Add(-time.Minute)}
	return req
}

// Each target is the base backup the entrypoint fetches and the recovery
// settings it writes (design §4, "What to restore to").
func TestEachTargetIsItsRecoverySettings(t *testing.T) {
	for kind, want := range map[string][3]string{
		entity.ManagedRestoreToLatest: {"LATEST", "", ""},
		entity.ManagedRestoreToTime: {"", "2026-10-18T14:32:00Z",
			"recovery_target_time=2026-10-18T14:32:00+00:00"},
		entity.ManagedRestoreToPoint: {"", "2026-10-18T14:31:00Z",
			"recovery_target_lsn=0/3000060 recovery_target_inclusive=on"},
		entity.ManagedRestoreToBackup: {"base_000000010000000000000004", "", "recovery_target=immediate"},
	} {
		req := restoreReq(t, entity.ManagedRestoreTarget{Kind: kind, Time: targetTime, Point: "before-v42",
			Position: "0/3000060", Backup: "base_000000010000000000000004"})

		assert.NoError(t, New().PrepareDoc(req), kind)

		env := envOf(req.Doc)
		assert.Equal(t, want[0], env[envRestoreBackup], kind)
		assert.Equal(t, want[1], env[envRestoreBefore], kind)
		assert.Equal(t, want[2], env[envRestoreRecovery], kind)
	}
}

// The source's archive is only read: its settings are under the restore's
// prefix, its keys in the restore's directory, and what the new app archives
// with is its own.
func TestARestoredAppReadsTheSourceAndArchivesToItsOwn(t *testing.T) {
	req := restoreReq(t, entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest})

	assert.NoError(t, New().PrepareDoc(req))

	env := envOf(req.Doc)
	assert.Equal(t, "s3://backups/pg/walg/shop-db-restored", env["WALG_S3_PREFIX"], "its own")
	assert.Equal(t, "s3://backups/pg/walg/shop-db", env[restoreEnvPrefix+"WALG_S3_PREFIX"], "the source's")
	assert.Equal(t, "https://s3.example.com", env[restoreEnvPrefix+"AWS_ENDPOINT"])
	assert.Equal(t, restoreSecretsDir, env[envRestoreSecretsDir])
	for _, value := range env {
		assert.NotContains(t, value, "source-secret-value", "no key in the environment")
		assert.NotContains(t, value, "source-libsodium-key")
	}

	secrets := secretsOf(req.Doc)
	for name, want := range map[string][2]string{
		"WALG_AWS_ACCESS_KEY_ID":                       {"AKIAEXAMPLE", secretsDir + "/WALG_AWS_ACCESS_KEY_ID"},
		"WALG_LIBSODIUM_KEY":                           {"libsodium-key-value", secretsDir + "/WALG_LIBSODIUM_KEY"},
		restoreSecretPrefix + "WALG_AWS_ACCESS_KEY_ID": {"AKIA-SOURCE", restoreSecretsDir + "/WALG_AWS_ACCESS_KEY_ID"},
		restoreSecretPrefix + "WALG_LIBSODIUM_KEY":     {"source-libsodium-key", restoreSecretsDir + "/WALG_LIBSODIUM_KEY"},
	} {
		if assert.Contains(t, secrets, name) {
			assert.Equal(t, want[0], secrets[name]["value"], name)
			file := secrets[name]["swarmRef"].(map[string]any)["file"].(map[string]any)
			assert.Equal(t, want[1], file["name"], name)
		}
	}
	assert.Contains(t, req.OwnedEnv, envRestoreBackup, "locked while it restores")
	assert.Contains(t, req.OwnedEnv, restoreEnvPrefix+"WALG_S3_PREFIX")
	assert.NoError(t, specmodel.CheckBuildable(req.Doc))
}

func TestARestoreFromAVolumeMountsTheSourceReadOnly(t *testing.T) {
	req := restoreReq(t, entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest})
	req.Archive = &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{VolumeID: "vol-repo",
		Subpath: "backups/walg/shop-db-restored"}}
	req.Restore.Archive = &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{VolumeID: "vol-repo",
		Subpath: "backups/walg/shop-db"}}

	assert.NoError(t, New().PrepareDoc(req))

	mounts := req.Doc.Deployment.Storage.Mounts
	assert.Equal(t, "backups/walg/shop-db-restored", mounts[volumeArchiveDir].VolumeOptions.Subpath)
	source := mounts[restoreSourceDir]
	assert.Equal(t, "vol-repo", source.Source)
	assert.Equal(t, "backups/walg/shop-db", source.VolumeOptions.Subpath)
	assert.True(t, source.ReadOnly, "the source is only read")
	env := envOf(req.Doc)
	assert.Equal(t, volumeArchiveDir, env["WALG_FILE_PREFIX"])
	assert.Equal(t, restoreSourceDir, env[restoreEnvPrefix+"WALG_FILE_PREFIX"])
	assert.Equal(t, []string{restoreSecretPrefix + "WALG_LIBSODIUM_KEY"}, restoreSecretNames(secretsOf(req.Doc)))
	assert.NoError(t, specmodel.CheckBuildable(req.Doc))
}

func restoreSecretNames(secrets map[string]map[string]any) []string {
	var out []string
	for name := range secrets {
		if strings.HasPrefix(name, restoreSecretPrefix) {
			out = append(out, name)
		}
	}
	return out
}

// A restored app starts through the engine's entrypoint, and says where it came
// from.
func TestARestoredAppStartsThroughTheEntrypoint(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)
	target := entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToTime, Time: targetTime}
	req.Prepared.Restore = &managedmodel.RestoreSource{SourceAppID: "app-src", Archive: sourceArchive(),
		Target: target}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	settings := deployment.MustAsAppDeploymentSettings()
	assert.Equal(t, scriptDir+"/walg-entrypoint", settings.Entrypoint)
	assert.True(t, strings.HasPrefix(settings.Command, "postgres "), "the server's arguments, as always")
	managed := resp.Settings[0].MustAsAppManagedSettings()
	assert.Equal(t, &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"}, Target: target},
		managed.Database.Restore)
	for _, job := range resp.Settings[1:] {
		assert.Equal(t, base.SettingStatusDisabled, job.Status, "%s waits for the recovery", job.Name)
	}
}

// A new app starts through the engine's entrypoint too: it hands over to the
// image's, with a checkpoint once the server is up, so that archive_timeout holds
// from the first minute.
func TestANewAppStartsThroughTheEntrypoint(t *testing.T) {
	deployment := deploymentSetting()

	resp, err := New().Configure(configureReq(t, deployment))

	assert.NoError(t, err)
	assert.Equal(t, scriptDir+"/walg-entrypoint", deployment.MustAsAppDeploymentSettings().Entrypoint)
	assert.Nil(t, resp.Settings[0].MustAsAppManagedSettings().Database.Restore)
	for _, job := range resp.Settings[1:] {
		assert.Equal(t, base.SettingStatusActive, job.Status, job.Name)
	}
}

// Once ready, a restored app keeps nothing of the source: its variables, its
// keys and its mount go, and so do their locks.
func TestFinishingARestoreDropsWhatItWasRestoredWith(t *testing.T) {
	env := &entity.Setting{ID: "env-1", Type: base.SettingTypeEnvVar, UpdateVer: 1}
	env.MustSetData(&entity.EnvVars{Data: []*entity.EnvVar{{Key: "WALG_S3_PREFIX", Value: "s3://b/new"},
		{Key: restoreEnvPrefix + "WALG_S3_PREFIX", Value: "s3://b/src"}, {Key: envRestoreRecovery, Value: "x"},
		{Key: "TZ", Value: "UTC"}}})
	ownKey := &entity.Setting{ID: "sec-1", Type: base.SettingTypeSecret, Name: "WALG_LIBSODIUM_KEY"}
	ownKey.MustSetData(&entity.Secret{Key: "WALG_LIBSODIUM_KEY"})
	sourceKey := &entity.Setting{ID: "sec-2", Type: base.SettingTypeSecret,
		Name: restoreSecretPrefix + "WALG_LIBSODIUM_KEY"}
	sourceKey.MustSetData(&entity.Secret{Key: restoreSecretPrefix + "WALG_LIBSODIUM_KEY"})
	ownMount := &entity.Setting{ID: "mnt-1", Type: base.SettingTypeAppSettingMount}
	ownMount.MustSetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: "sec-1"}})
	sourceMount := &entity.Setting{ID: "mnt-2", Type: base.SettingTypeAppSettingMount}
	sourceMount.MustSetData(&entity.AppSettingMount{Source: entity.ObjectID{ID: "sec-2"}})
	now := time.Date(2026, 10, 18, 16, 0, 0, 0, time.UTC)
	managed := &entity.AppManagedSettings{Engine: "postgres",
		Locks: entity.ManagedLocks{EnvKeys: []string{"WALG_S3_PREFIX", restoreEnvPrefix + "WALG_S3_PREFIX",
			envRestoreRecovery}, Mounts: []string{"/var/lib/postgresql", restoreSourceDir},
			Settings: []string{"sec-1", "sec-2", "mnt-1", "mnt-2"}},
		Database: &entity.ManagedDatabase{Restore: &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"}}}}

	resp, err := New().FinishRestore(&managedmodel.FinishRestoreReq{Managed: managed,
		Settings: []*entity.Setting{env, ownKey, sourceKey, ownMount, sourceMount}, Now: now})

	assert.NoError(t, err)
	if assert.Len(t, resp.Changed, 1) {
		kept := []string{}
		for _, v := range resp.Changed[0].MustAsEnvVars().Data {
			kept = append(kept, v.Key)
		}
		assert.Equal(t, []string{"WALG_S3_PREFIX", "TZ"}, kept)
		assert.Equal(t, 2, resp.Changed[0].UpdateVer)
	}
	assert.ElementsMatch(t, []string{"sec-2", "mnt-2"}, ids(resp.Deleted))
	assert.ElementsMatch(t, []string{restoreEnvPrefix + "WALG_S3_PREFIX", envRestoreRecovery}, resp.EnvKeys)
	assert.Equal(t, []string{restoreSourceDir}, resp.MountTargets)
	assert.Equal(t, []string{"WALG_S3_PREFIX"}, managed.Locks.EnvKeys)
	assert.Equal(t, []string{"/var/lib/postgresql"}, managed.Locks.Mounts)
	assert.Equal(t, []string{"sec-1", "mnt-1"}, managed.Locks.Settings)
	assert.Equal(t, now, managed.Database.Restore.ReadyAt)
	assert.Equal(t, "app-src", managed.Database.Restore.SourceApp.ID, "where it came from stays told")
}

func TestTheFirstBackupIsTheNightlyJobsRun(t *testing.T) {
	managed, _ := reconfigureSetup(t)

	assert.Equal(t, managed.Database.Jobs.Nightly.ID, New().FirstBackupJob(managed))
	assert.Equal(t, scriptDir+"/walg-entrypoint status", New().RecoveryCheck().Command)
}
