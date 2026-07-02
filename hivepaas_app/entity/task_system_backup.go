package entity

type TaskSystemBackupOutput struct {
	DBBackup *DBBackupOutput `json:"dbBackup"`
}

type DBBackupOutput struct {
	Error string `json:"error,omitempty"`
}

func (t *Task) OutputAsSystemBackup() (*TaskSystemBackupOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSystemBackupOutput { return &TaskSystemBackupOutput{} })
}
