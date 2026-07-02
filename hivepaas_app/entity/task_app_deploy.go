package entity

type TaskAppDeployArgs struct {
	Deployment ObjectID `json:"deployment"`
}

type TaskAppDeployOutput struct {
}

func (t *Task) ArgsAsAppDeploy() (*TaskAppDeployArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppDeployArgs { return &TaskAppDeployArgs{} })
}

func (t *Task) OutputAsAppDeploy() (*TaskAppDeployOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppDeployOutput { return &TaskAppDeployOutput{} })
}
