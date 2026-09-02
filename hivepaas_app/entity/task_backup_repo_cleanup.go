package entity

type TaskBackupRepoCleanupArgs struct {
	TargetRepos ObjectIDSlice `json:"targetRepos"`
}

type TaskBackupRepoCleanupOutput struct {
	// ReposCleaned / ReposSkipped / ReposFailed summarize the run. Skipped means another cleanup
	// already held the repository's lock, which is an expected outcome rather than a failure.
	ReposCleaned int `json:"reposCleaned"`
	ReposSkipped int `json:"reposSkipped"`
	ReposFailed  int `json:"reposFailed"`

	// Repos carries the per-repository outcome so the run is auditable rather than a bare count.
	Repos []*TaskBackupRepoCleanupRepoOutput `json:"repos,omitempty"`
}

type TaskBackupRepoCleanupRepoOutput struct {
	RepoID   string `json:"repoId"`
	RepoName string `json:"repoName"`

	// Skipped is set when the repository was already being cleaned up elsewhere.
	Skipped bool `json:"skipped,omitempty"`
	// Error is set when this repository failed; the others still run.
	Error string `json:"error,omitempty"`

	SnapshotsInRepo int `json:"snapshotsInRepo"`
	RecordsRemoved  int `json:"recordsRemoved"`
	RecordsAdded    int `json:"recordsAdded"`
}

func (t *Task) ArgsAsBackupRepoCleanup() (*TaskBackupRepoCleanupArgs, error) {
	return parseTaskArgsAs(t, func() *TaskBackupRepoCleanupArgs { return &TaskBackupRepoCleanupArgs{} })
}

func (t *Task) OutputAsBackupRepoCleanup() (*TaskBackupRepoCleanupOutput, error) {
	return parseTaskOutputAs(t, func() *TaskBackupRepoCleanupOutput { return &TaskBackupRepoCleanupOutput{} })
}
