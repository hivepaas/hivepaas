package entity

// TaskAppLabelsSweepArgs carries what the sweep needs to reach every app.
//
// Only the HivePaaS app id: the settings themselves are read fresh when the task
// runs, deliberately. The task is scheduled when a change is confirmed and may
// run seconds later or, after a retry, minutes later - and what should be on the
// apps by then is whatever the settings say at that moment, not what they said
// when somebody pressed a button.
type TaskAppLabelsSweepArgs struct {
	AppID string `json:"appId"`
}

type TaskAppLabelsSweepOutput struct {
	Applied int               `json:"applied"`
	Skipped int               `json:"skipped"`
	Failed  map[string]string `json:"failed,omitempty"`
}

func (t *Task) ArgsAsAppLabelsSweep() (*TaskAppLabelsSweepArgs, error) {
	return parseTaskArgsAs(t, func() *TaskAppLabelsSweepArgs { return &TaskAppLabelsSweepArgs{} })
}

func (t *Task) OutputAsAppLabelsSweep() (*TaskAppLabelsSweepOutput, error) {
	return parseTaskOutputAs(t, func() *TaskAppLabelsSweepOutput { return &TaskAppLabelsSweepOutput{} })
}
