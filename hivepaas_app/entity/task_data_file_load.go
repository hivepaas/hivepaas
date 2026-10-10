package entity

// TaskDataFileLoadArgs is a load of a data file into a command, as it was asked
// for. The task's target is the file, its object the app.
type TaskDataFileLoadArgs struct {
	ProjectID string `json:"projectId"`
	AppID     string `json:"appId"`
	// Command reads the file on its stdin.
	Command *CommandTemplate `json:"command"`
	// Passphrase decrypts a file saved encrypted; none for one that is not.
	Passphrase EncryptedField `json:"passphrase"`
}

func (t *Task) ArgsAsDataFileLoad() (*TaskDataFileLoadArgs, error) {
	return parseTaskArgsAs(t, func() *TaskDataFileLoadArgs { return &TaskDataFileLoadArgs{} })
}
