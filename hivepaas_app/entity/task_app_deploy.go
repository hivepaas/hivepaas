package entity

type TaskAppDeployArgs struct {
	Deployment ObjectID `json:"deployment"`

	// NoCache and ImageTags are what this one deployment asked for. They are here
	// rather than in the app's deployment settings because that is what they are:
	// arguments to one run, not configuration of the app.
	NoCache bool `json:"noCache,omitempty"`
	// ImageTags carry no environment prefix; the build adds it.
	ImageTags []string `json:"imageTags,omitempty"`
}

type TaskAppDeployOutput struct {
}

func (t *Task) ArgsAsAppDeploy() (*TaskAppDeployArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppDeployArgs { return &TaskAppDeployArgs{} })
}

func (t *Task) OutputAsAppDeploy() (*TaskAppDeployOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppDeployOutput { return &TaskAppDeployOutput{} })
}
