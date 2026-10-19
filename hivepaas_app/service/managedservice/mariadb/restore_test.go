package mariadb

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

var targetTime = time.Date(2026, 10, 19, 14, 32, 0, 0, time.UTC)

func sourceArchive() *managedmodel.Archive {
	return &managedmodel.Archive{S3: &managedmodel.S3Archive{
		Endpoint: "https://s3.example.com", Region: "auto", Bucket: "backups", Prefix: "walg/shop-db/app-1",
		AccessKey: "AKIA-SOURCE", SecretKey: "source-secret-value", ForcePathStyle: true,
	}}
}

// restoreReq is a restore of shop-db into a new app, which archives to its own
// prefix of the same bucket.
func restoreReq(t *testing.T, target entity.ManagedRestoreTarget) *managedmodel.PrepareReq {
	t.Helper()
	own := s3Archive()
	own.S3.Prefix = "walg/shop-db-restored/app-2"
	req := prepareReq(t, own)
	req.Restore = &managedmodel.RestoreSource{SourceAppID: "app-1", Archive: sourceArchive(),
		EncryptionKey: "source-libsodium-key", Target: target, PointAt: targetTime.Add(-time.Minute)}
	return req
}

// Each target is the base backup walg-entrypoint fetches and where its replay of
// the binary logs stops: the last one archived, a time, a point's position, or
// nowhere - the base backup as it is.
func TestEachTargetIsWhereTheReplayStops(t *testing.T) {
	for kind, want := range map[string]map[string]string{
		entity.ManagedRestoreToLatest: {walgkit.EnvRestoreBackup: "LATEST"},
		entity.ManagedRestoreToTime: {walgkit.EnvRestoreBefore: "2026-10-19T14:32:00Z",
			envRestoreStopTime: "2026-10-19T14:32:00Z"},
		entity.ManagedRestoreToPoint: {walgkit.EnvRestoreBefore: "2026-10-19T14:31:00Z",
			envRestoreStopPosition: "mariadb-bin.000003:389"},
		entity.ManagedRestoreToBackup: {walgkit.EnvRestoreBackup: "stream_20261019T020001Z",
			envRestoreReplay: "none"},
	} {
		req := restoreReq(t, entity.ManagedRestoreTarget{Kind: kind, Time: targetTime, Point: "before-v42",
			Position: "mariadb-bin.000003:389", Backup: "stream_20261019T020001Z"})

		assert.NoError(t, New().PrepareDoc(req), kind)

		env := envOf(req.Doc)
		for _, key := range []string{walgkit.EnvRestoreBackup, walgkit.EnvRestoreBefore, envRestoreStopTime,
			envRestoreStopPosition, envRestoreReplay} {
			assert.Equal(t, want[key], env[key], "%s: %s", kind, key)
		}
	}
}

// What would reach walg-entrypoint's arguments or the replay is held to what it
// may be.
func TestATargetThatCannotBeRecoveredToIsRefused(t *testing.T) {
	for name, target := range map[string]entity.ManagedRestoreTarget{
		"no time":            {Kind: entity.ManagedRestoreToTime},
		"a Postgres LSN":     {Kind: entity.ManagedRestoreToPoint, Position: "0/3000060"},
		"a position and SQL": {Kind: entity.ManagedRestoreToPoint, Position: "mariadb-bin.000003:389; drop"},
		"a Postgres backup":  {Kind: entity.ManagedRestoreToBackup, Backup: "base_000000010000000000000004"},
		"an unknown kind":    {Kind: "yesterday"},
	} {
		err := New().PrepareDoc(restoreReq(t, target))
		assert.ErrorIs(t, err, hperrors.ErrAppManagedRestoreTargetInvalid, name)
	}
}

// The source's archive is only read: its settings are under the restore's
// prefix, its keys in the restore's directory, readable by the mysql user, and
// what the new app archives with is its own.
func TestARestoredAppReadsTheSourceAndArchivesToItsOwn(t *testing.T) {
	req := restoreReq(t, entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest})

	assert.NoError(t, New().PrepareDoc(req))

	env := envOf(req.Doc)
	assert.Equal(t, "s3://backups/walg/shop-db-restored/app-2", env["WALG_S3_PREFIX"], "its own")
	assert.Equal(t, "s3://backups/walg/shop-db/app-1", env[walgkit.RestoreEnvPrefix+"WALG_S3_PREFIX"], "the source's")
	assert.Equal(t, walgkit.RestoreSecretsDir, env[walgkit.EnvRestoreSecretsDir])
	for _, value := range env {
		assert.NotContains(t, value, "source-secret-value", "no key in the environment")
		assert.NotContains(t, value, "source-libsodium-key")
	}
	secrets := secretsOf(req.Doc)
	source := secrets[walgkit.RestoreSecretPrefix+"WALG_LIBSODIUM_KEY"]
	if assert.NotNil(t, source) {
		assert.Equal(t, "source-libsodium-key", source["value"])
		file := source["swarmRef"].(map[string]any)["file"].(map[string]any)
		assert.Equal(t, walgkit.RestoreSecretsDir+"/WALG_LIBSODIUM_KEY", file["name"])
		assert.Equal(t, mysqlUID, file["uid"], "walg-run --source runs as mysql")
	}
	assert.Contains(t, req.OwnedEnv, walgkit.EnvRestoreBackup, "locked while it restores")
	assert.NoError(t, specmodel.CheckBuildable(req.Doc))
}

func TestARestoreFromAVolumeMountsTheSourceReadOnly(t *testing.T) {
	req := restoreReq(t, entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest})
	req.Restore.Archive = &managedmodel.Archive{Volume: &managedmodel.VolumeArchive{VolumeID: "vol-repo",
		Subpath: "backups/walg/shop-db"}}

	assert.NoError(t, New().PrepareDoc(req))

	source := req.Doc.Deployment.Storage.Mounts[walgkit.RestoreSourceDir]
	assert.Equal(t, "vol-repo", source.Source)
	assert.True(t, source.ReadOnly, "the source is only read")
	assert.Equal(t, walgkit.RestoreSourceDir, envOf(req.Doc)[walgkit.RestoreEnvPrefix+"WALG_FILE_PREFIX"])
}

// A restored app starts through the entrypoint, its jobs waiting for the replay,
// and says where it came from.
func TestARestoredAppStartsThroughTheEntrypoint(t *testing.T) {
	deployment := deploymentSetting()
	req := configureReq(t, deployment)
	target := entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToTime, Time: targetTime}
	req.Prepared.Restore = &managedmodel.RestoreSource{SourceAppID: "app-src", Archive: sourceArchive(),
		Target: target}

	resp, err := New().Configure(req)

	assert.NoError(t, err)
	settings := deployment.MustAsAppDeploymentSettings()
	assert.Equal(t, walgkit.ScriptDir+"/walg-entrypoint", settings.Entrypoint)
	assert.True(t, strings.HasPrefix(settings.Command, "mariadbd "))
	managed := resp.Settings[0].MustAsAppManagedSettings()
	assert.Equal(t, &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"}, Target: target},
		managed.Database.Restore)
	for _, job := range resp.Settings[1:] {
		assert.Equal(t, base.SettingStatusDisabled, job.Status, "%s waits for the replay", job.Name)
	}
}

// The restore asks walg-entrypoint whether the replay ended, takes the first base
// backup with the nightly job, then takes off the app what it was restored with.
func TestARestoreIsSeenThrough(t *testing.T) {
	engine := New()
	assert.Equal(t, walgkit.ScriptDir+"/walg-entrypoint status", engine.RecoveryCheck().Command)

	managed := &entity.AppManagedSettings{Engine: "mariadb", EngineAPI: 1, Database: &entity.ManagedDatabase{
		Jobs:    entity.ManagedJobs{Nightly: entity.ObjectID{ID: "nightly"}},
		Restore: &entity.ManagedRestore{SourceApp: entity.ObjectID{ID: "app-src"}},
	}, Locks: entity.ManagedLocks{EnvKeys: []string{"WALG_S3_PREFIX", walgkit.EnvRestoreBackup, envRestoreStopTime}}}
	assert.Equal(t, "nightly", engine.FirstBackupJob(managed))

	env := &entity.Setting{ID: "env-1", Type: base.SettingTypeEnvVar}
	env.MustSetData(&entity.EnvVars{Data: []*entity.EnvVar{{Key: "WALG_S3_PREFIX", Value: "s3://b/p"},
		{Key: walgkit.EnvRestoreBackup, Value: "LATEST"}, {Key: envRestoreStopTime, Value: "2026-10-19T14:32:00Z"}}})
	secret := &entity.Setting{ID: "secret-1", Type: base.SettingTypeSecret, Name: walgkit.RestoreSecretPrefix + "X"}

	resp, err := engine.FinishRestore(&managedmodel.FinishRestoreReq{Managed: managed,
		Settings: []*entity.Setting{env, secret}, Now: targetTime})

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{walgkit.EnvRestoreBackup, envRestoreStopTime}, resp.EnvKeys)
	assert.Equal(t, []*entity.Setting{secret}, resp.Deleted)
	assert.Equal(t, []string{"WALG_S3_PREFIX"}, managed.Locks.EnvKeys)
	assert.Equal(t, targetTime, managed.Database.Restore.ReadyAt)
}
