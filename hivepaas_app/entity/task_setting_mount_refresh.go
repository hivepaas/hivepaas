package entity

// TaskSettingMountRefreshArgs are the apps whose mounted settings to bring up to
// date. Only their ids: what they mount is read when the task runs, which may be
// minutes later after a retry.
type TaskSettingMountRefreshArgs struct {
	AppIDs []string `json:"appIds"`
}

type TaskSettingMountRefreshOutput struct {
	Applied int               `json:"applied"`
	Failed  map[string]string `json:"failed,omitempty"`
}

func (t *Task) ArgsAsSettingMountRefresh() (*TaskSettingMountRefreshArgs, error) {
	return parseTaskArgsAs(t, func() *TaskSettingMountRefreshArgs { return &TaskSettingMountRefreshArgs{} })
}

func (t *Task) OutputAsSettingMountRefresh() (*TaskSettingMountRefreshOutput, error) {
	return parseTaskOutputAs(t, func() *TaskSettingMountRefreshOutput { return &TaskSettingMountRefreshOutput{} })
}
