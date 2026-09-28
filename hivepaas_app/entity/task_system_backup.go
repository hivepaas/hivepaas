package entity

// TaskSystemBackupOutput is what a system backup's run took. Its snapshot's
// fields are named as a data backup's are, so that a run's page reads it the
// same way.
type TaskSystemBackupOutput struct {
	SnapshotID string `json:"snapshotId"`
	SizeBytes  int64  `json:"sizeBytes"`
	// Includes is what the snapshot holds: database, spec.
	Includes []string `json:"includes,omitempty"`
}

func (t *Task) OutputAsSystemBackup() (*TaskSystemBackupOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSystemBackupOutput { return &TaskSystemBackupOutput{} })
}
