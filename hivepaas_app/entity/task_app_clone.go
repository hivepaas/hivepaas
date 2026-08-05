package entity

type TaskAppCloneArgs struct {
	SrcApp ObjectID `json:"srcApp"`
}

type TaskAppCloneOutput struct {
}

func (t *Task) ArgsAsAppClone() (*TaskAppCloneArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppCloneArgs { return &TaskAppCloneArgs{} })
}

func (t *Task) OutputAsAppClone() (*TaskAppCloneOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppCloneOutput { return &TaskAppCloneOutput{} })
}
