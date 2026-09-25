package entity

type TaskAppCloneArgs struct {
	SrcApp ObjectID `json:"srcApp"`
	// DropGatedMounts is the answer of the Reveal Secrets gate at the request:
	// setting mounts with a gated part are not copied.
	DropGatedMounts bool `json:"dropGatedMounts,omitempty"`
}

type TaskAppCloneOutput struct {
}

func (t *Task) ArgsAsAppClone() (*TaskAppCloneArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppCloneArgs { return &TaskAppCloneArgs{} })
}

func (t *Task) OutputAsAppClone() (*TaskAppCloneOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppCloneOutput { return &TaskAppCloneOutput{} })
}
