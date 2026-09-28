package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// A data backup references the repository it writes to and the volume it reads.
func TestSchedJobDataBackupReferences(t *testing.T) {
	job := &SchedJob{
		JobType: base.SchedJobTypeDataBackup,
		App:     ObjectID{ID: "a1"},
		DataBackup: &SchedJobDataBackup{
			Source:           base.SchedJobDataBackupSourceVolume,
			SourceVolume:     ObjectID{ID: "vol1"},
			TargetRepository: ObjectID{ID: "repo1"},
		},
	}

	refs := job.GetRefObjectIDs()

	assert.ElementsMatch(t, []string{"vol1", "repo1"}, refs.RefSettingIDs)
	assert.Equal(t, []string{"a1"}, refs.RefAppIDs)
}

// A command's backup has no volume to reference.
func TestSchedJobDataBackupOfACommandReferencesTheRepositoryOnly(t *testing.T) {
	job := &SchedJob{JobType: base.SchedJobTypeDataBackup, DataBackup: &SchedJobDataBackup{
		Source:           base.SchedJobDataBackupSourceCommand,
		SourceCommand:    &CommandTemplate{Command: "pg_dump app"},
		SourceFileName:   "db.sql",
		TargetRepository: ObjectID{ID: "repo1"},
	}}

	assert.Equal(t, []string{"repo1"}, job.GetRefObjectIDs().RefSettingIDs)
}

// A snapshot's tags carry the job, the app and the source, then the job's own,
// as kopia takes them: key:value, in a stable order.
// The scripts its commands run are references too: they keep a script from
// being deleted under the job.
func TestSchedJobDataBackupReferencesItsCommandsScripts(t *testing.T) {
	job := &SchedJob{JobType: base.SchedJobTypeDataBackup, DataBackup: &SchedJobDataBackup{
		Source:           base.SchedJobDataBackupSourceCommand,
		SourceCommand:    &CommandTemplate{Script: ObjectValue{ID: "dump-script"}},
		RestoreCommand:   &CommandTemplate{Script: ObjectValue{ID: "load-script"}},
		SourceFileName:   "db.sql",
		TargetRepository: ObjectID{ID: "repo1"},
	}}

	assert.ElementsMatch(t, []string{"dump-script", "load-script", "repo1"}, job.GetRefObjectIDs().RefSettingIDs)
}

func TestSchedJobDataBackupSnapshotTags(t *testing.T) {
	backup := &SchedJobDataBackup{Source: base.SchedJobDataBackupSourceCommand,
		Tags: map[string]string{"env": "prod", "db": "main"}}

	assert.Equal(t, []string{"hivepaas.job:j1", "hivepaas.app:a1", "hivepaas.source:command", "db:main", "env:prod"},
		backup.SnapshotTags("j1", "a1"))
}

// A snapshot's tags read back: which app, job, run and source it came from.
func TestParseDataBackupSnapshotTags(t *testing.T) {
	parsed := ParseDataBackupSnapshotTags([]string{
		"hivepaas.job:j1", "hivepaas.app:a1", "hivepaas.run:t1", "hivepaas.source:volume", "env:prod",
	})

	assert.Equal(t, &DataBackupSnapshotTags{AppID: "a1", JobID: "j1", RunID: "t1",
		Source: base.SchedJobDataBackupSourceVolume}, parsed)
	assert.Equal(t, &DataBackupSnapshotTags{}, ParseDataBackupSnapshotTags([]string{"env:prod"}))
}

// A run's result is its task's output; a task that is no data backup's has none.
func TestTaskOutputAsDataBackup(t *testing.T) {
	task := &Task{Type: base.TaskTypeSchedJobExec}
	result, err := task.OutputAsDataBackup()
	assert.NoError(t, err)
	assert.Nil(t, result)

	task.MustSetOutput(&SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42})
	reread := &Task{Type: base.TaskTypeSchedJobExec, Output: task.Output}
	result, err = reread.OutputAsDataBackup()
	assert.NoError(t, err)
	assert.Equal(t, &SchedJobDataBackupResult{SnapshotID: "k1", SizeBytes: 42}, result)

	sequenceRun := &Task{Type: base.TaskTypeSchedJobExec}
	sequenceRun.MustSetOutput(&SchedJobSeqRun{Started: true})
	reread = &Task{Type: base.TaskTypeSchedJobExec, Output: sequenceRun.Output}
	result, err = reread.OutputAsDataBackup()
	assert.NoError(t, err)
	assert.Nil(t, result, "a sequence's run is no backup's result")

	backupRun := &Task{Type: base.TaskTypeSchedJobExec}
	backupRun.MustSetOutput(&SchedJobDataBackupResult{SnapshotID: "k1"})
	reread = &Task{Type: base.TaskTypeSchedJobExec, Output: backupRun.Output}
	_, _ = reread.OutputAsSchedJobSeqRun()
	result, err = reread.OutputAsDataBackup()
	assert.NoError(t, err, "read after the task's output was parsed as a sequence's run")
	assert.Equal(t, "k1", result.SnapshotID)
}
