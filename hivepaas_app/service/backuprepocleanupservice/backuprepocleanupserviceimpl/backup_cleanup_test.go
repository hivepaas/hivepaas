package backuprepocleanupserviceimpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// The task decides what to clean: named repositories, or all of them when it names none. Getting
// this wrong either prunes repositories nobody asked for, or silently prunes nothing.
func TestTaskArgs_EmptyMeansEveryRepo(t *testing.T) {
	task := &entity.Task{}

	args, err := task.ArgsAsBackupRepoCleanup()
	assert.NoError(t, err)
	// A task with no args parses to nil, which loadTargetRepos treats as "no filter".
	assert.Nil(t, args)
}

func TestTaskArgs_RoundTripsTargetRepos(t *testing.T) {
	task := &entity.Task{}
	task.MustSetArgs(&entity.TaskBackupRepoCleanupArgs{
		TargetRepos: entity.ObjectIDSlice{{ID: "repo-a"}, {ID: "repo-b"}},
	})

	args, err := task.ArgsAsBackupRepoCleanup()
	assert.NoError(t, err)
	assert.Len(t, args.TargetRepos, 2)
	assert.Equal(t, "repo-a", args.TargetRepos[0].ID)
}

// The output is what the user sees after a scheduled run, so the counters and the per-repo detail
// have to survive the round trip through the task record.
func TestTaskOutput_RoundTrips(t *testing.T) {
	task := &entity.Task{}
	task.MustSetOutput(&entity.TaskBackupRepoCleanupOutput{
		ReposCleaned: 2,
		ReposSkipped: 1,
		ReposFailed:  1,
		Repos: []*entity.TaskBackupRepoCleanupRepoOutput{
			{RepoID: "repo-a", RepoName: "A", SnapshotsInRepo: 5, RecordsRemoved: 3, RecordsAdded: 1},
			{RepoID: "repo-b", RepoName: "B", Skipped: true},
			{RepoID: "repo-c", RepoName: "C", Error: "storage unreachable"},
		},
	})

	out, err := task.OutputAsBackupRepoCleanup()
	assert.NoError(t, err)
	assert.Equal(t, 2, out.ReposCleaned)
	assert.Equal(t, 1, out.ReposSkipped)
	assert.Equal(t, 1, out.ReposFailed)
	assert.Len(t, out.Repos, 3)

	assert.Equal(t, 3, out.Repos[0].RecordsRemoved)
	// A skipped repository is not a failure: another cleanup already had it.
	assert.True(t, out.Repos[1].Skipped)
	assert.Empty(t, out.Repos[1].Error)
	assert.Equal(t, "storage unreachable", out.Repos[2].Error)
}
