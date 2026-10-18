package entity

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// TaskSchedJobOutputs is what a managed job's run wrote to its output file, as
// KEY=value lines: what its engine reads the app's state from.
type TaskSchedJobOutputs struct {
	Outputs map[string]string `json:"outputs"`
}

// OutputAsSchedJobOutputs is a managed job's run's outputs; nil for a task that
// holds none, another job's among them.
func (t *Task) OutputAsSchedJobOutputs() (*TaskSchedJobOutputs, error) {
	if t.Output == "" {
		return nil, nil //nolint:nilnil // no output, no outputs
	}
	// Read apart from the task's parsed output, as a data backup's is.
	result := &TaskSchedJobOutputs{}
	if err := json.Unmarshal([]byte(t.Output), result); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if result.Outputs == nil {
		return nil, nil //nolint:nilnil // no outputs: another job's output
	}
	return result, nil
}
