package entity

type TaskAppPreviewArgs struct {
	ParentApp       ObjectID              `json:"parentApp"`
	RepoRef         string                `json:"repoRef"`
	CustomSubdomain string                `json:"customSubdomain,omitempty"`
	NoStart         bool                  `json:"noStart,omitempty"`
	CloneDBApps     bool                  `json:"cloneDbApps,omitempty"`
	Trigger         *AppDeploymentTrigger `json:"trigger,omitempty"`
}

type TaskAppPreviewOutput struct {
	PreviewApp   ObjectID      `json:"previewApp,omitempty"`
	Deployment   ObjectID      `json:"deployment,omitempty"`
	ClonedDBApps ObjectIDSlice `json:"clonedDbApps,omitempty"`
}

func (t *Task) ArgsAsAppPreview() (*TaskAppPreviewArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppPreviewArgs { return &TaskAppPreviewArgs{} })
}

func (t *Task) OutputAsAppPreview() (*TaskAppPreviewOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppPreviewOutput { return &TaskAppPreviewOutput{} })
}
