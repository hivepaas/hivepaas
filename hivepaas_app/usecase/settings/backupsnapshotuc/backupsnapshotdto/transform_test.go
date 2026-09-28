package backupsnapshotdto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func snapshotRecord(t *testing.T, id string) *entity.Setting {
	t.Helper()
	s := &entity.Setting{ID: id, RefID: "r1", Type: base.SettingTypeBackupSnapshot, Name: "k1"}
	s.MustSetData(&entity.BackupSnapshot{ID: "k1full", ShortID: "k1", Description: "nightly (run t1)",
		Time: time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC), Paths: []string{"/j1"}, Hostname: "data-backup",
		SizeBytes: 42})
	return s
}

func refs() *SnapshotRefs {
	job := &entity.Setting{ID: "j1", Name: "nightly", Type: base.SettingTypeSchedJob}
	job.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeDataBackup,
		DataBackup: &entity.SchedJobDataBackup{Source: base.SchedJobDataBackupSourceVolume}})
	return &SnapshotRefs{
		Repos: map[string]*entity.Setting{"r1": {ID: "r1", Name: "s3", Type: base.SettingTypeBackupRepo,
			Status: base.SettingStatusActive}},
		Apps: map[string]*entity.App{"a1": {ID: "a1", Name: "web", ProjectEnvID: "p1:dev"}},
		Jobs: map[string]*entity.Setting{"j1": job},
	}
}

// A snapshot reads with its repository, its app, its job and its run named, from
// its tags; the source is its tag's, or its job's for a snapshot taken before
// the tag.
func TestTransformBackupSnapshot(t *testing.T) {
	resp := TransformBackupSnapshot(snapshotRecord(t, "s1"),
		[]string{"hivepaas.app:a1", "hivepaas.job:j1", "hivepaas.run:t1", "env:prod"}, refs())

	assert.Equal(t, "s1", resp.ID)
	assert.Equal(t, "k1full", resp.SnapshotID)
	assert.Equal(t, "k1", resp.ShortID)
	assert.Equal(t, int64(42), resp.SizeBytes)
	assert.Equal(t, "s3", resp.Repo.Name)
	assert.Equal(t, &SnapshotAppResp{ID: "a1", Name: "web", Env: "dev"}, resp.App)
	assert.Equal(t, "nightly", resp.Job.Name)
	assert.Equal(t, "t1", resp.RunID)
	assert.Equal(t, base.SchedJobDataBackupSourceVolume, resp.Source, "from the job: no source tag")
	assert.Equal(t, []string{"hivepaas.app:a1", "hivepaas.job:j1", "hivepaas.run:t1", "env:prod"}, resp.Tags)

	tagged := TransformBackupSnapshot(snapshotRecord(t, "s2"),
		[]string{"hivepaas.app:a1", "hivepaas.source:command"}, refs())
	assert.Equal(t, base.SchedJobDataBackupSourceCommand, tagged.Source)
}

// An app or a job that is gone is still named, as deleted; a snapshot no data
// backup took has neither.
func TestTransformBackupSnapshotOfWhatIsGone(t *testing.T) {
	resp := TransformBackupSnapshot(snapshotRecord(t, "s1"),
		[]string{"hivepaas.app:gone-app", "hivepaas.job:gone-job"}, refs())
	assert.Equal(t, &SnapshotAppResp{ID: "gone-app", Deleted: true}, resp.App)
	assert.Equal(t, &SnapshotJobResp{ID: "gone-job", Deleted: true}, resp.Job)
	assert.Empty(t, resp.Source)

	plain := TransformBackupSnapshot(snapshotRecord(t, "s2"), []string{"env:prod"}, refs())
	assert.Nil(t, plain.App)
	assert.Nil(t, plain.Job)
}
