package entity

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// SchedJobSeqRun is a job sequence's run: where it is and how each step went.
// It is the output of the run's task, written after every step, so the run's
// page shows it while the run goes on.
type SchedJobSeqRun struct {
	// Started is set once the run has checked it does not overlap another.
	Started     bool                     `json:"started"`
	CurrentStep int                      `json:"currentStep"`
	Steps       []*SchedJobSeqStepResult `json:"steps"`
}

type SchedJobSeqStepResult struct {
	Job       ObjectID                   `json:"job"`
	Name      string                     `json:"name,omitempty"`
	Status    base.SchedJobSeqStepStatus `json:"status"`
	ExitCode  *int                       `json:"exitCode,omitempty"`
	Error     string                     `json:"error,omitempty"`
	Attempts  int                        `json:"attempts"`
	StartedAt time.Time                  `json:"startedAt,omitzero"`
	EndedAt   time.Time                  `json:"endedAt,omitzero"`
	Outputs   map[string]string          `json:"outputs,omitempty"`
}

// NewSchedJobSeqRun is a run of the sequence with every step pending.
func NewSchedJobSeqRun(seq *SchedJobSequence) *SchedJobSeqRun {
	run := &SchedJobSeqRun{}
	if seq == nil {
		return run
	}
	for _, step := range seq.Steps {
		run.Steps = append(run.Steps, &SchedJobSeqStepResult{
			Job:    ObjectID{ID: step.Job.ID},
			Name:   step.Name,
			Status: base.SchedJobSeqStepPending,
		})
	}
	return run
}

// Failed says whether a step of the run failed.
func (r *SchedJobSeqRun) Failed() bool {
	for _, step := range r.Steps {
		if step.Status == base.SchedJobSeqStepFailed {
			return true
		}
	}
	return false
}

// OutputAsSchedJobSeqRun is the run a job sequence's task keeps in its output;
// nil for a task that holds none.
func (t *Task) OutputAsSchedJobSeqRun() (*SchedJobSeqRun, error) {
	if t.Type != base.TaskTypeSchedJobExec {
		return nil, hperrors.NewMismatch("Task type", base.TaskTypeSchedJobExec)
	}
	if t.Output == "" {
		return nil, nil //nolint:nilnil // a task that is not a sequence's holds none
	}
	return parseTaskOutputAs(t, func() *SchedJobSeqRun {
		return &SchedJobSeqRun{}
	})
}
