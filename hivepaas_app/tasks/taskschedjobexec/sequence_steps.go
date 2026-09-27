package taskschedjobexec

import (
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

// seqNext is what a job sequence's run does after a step.
type seqNext int

const (
	seqContinue seqNext = iota // run the next step
	seqDone                    // every step ran, none failed
	seqFailed                  // a step failed; with stop, the rest were skipped
)

const skippedAfterFailure = "an earlier step failed"

// stepConfig is the task config a step runs with: the member's retry and
// timeout - the sequence's timeout when the member sets none - under the run's
// own priority and control, with no retry spent yet.
func stepConfig(runConfig entity.TaskConfig, member, seq *entity.SchedJob) entity.TaskConfig {
	cfg := runConfig
	cfg.MaxRetry = member.MaxRetry
	cfg.Retry = 0
	cfg.RetryDelay = member.RetryDelay
	cfg.RetryDelayIncr = member.RetryDelayIncr
	cfg.RetryBackoffJitter = member.RetryBackoffJitter
	cfg.RetryDelayMax = member.RetryDelayMax
	cfg.Timeout = member.Timeout
	if cfg.Timeout <= 0 {
		cfg.Timeout = seq.Timeout
	}
	return cfg
}

// retriesLeft says whether a failed step is run again by the queue's retry.
func retriesLeft(cfg entity.TaskConfig) bool {
	return cfg.Retry < cfg.MaxRetry
}

// afterStep moves the run past its current step, which has ended, and says
// what comes next. With stop, a step that failed or was skipped ends the run
// and the steps after it are skipped.
func afterStep(run *entity.SchedJobSeqRun, stop bool) seqNext {
	step := run.Steps[run.CurrentStep]
	if stop && (step.Status == base.SchedJobSeqStepFailed || step.Status == base.SchedJobSeqStepSkipped) {
		skipFrom(run, run.CurrentStep+1, skippedAfterFailure)
		return seqFailed
	}
	run.CurrentStep++
	if run.CurrentStep < len(run.Steps) {
		return seqContinue
	}
	if run.Failed() {
		return seqFailed
	}
	return seqDone
}

// markCanceled ends a canceled run: the step it was on failed, the rest skipped.
func markCanceled(run *entity.SchedJobSeqRun) {
	if run.CurrentStep < len(run.Steps) {
		step := run.Steps[run.CurrentStep]
		step.Status = base.SchedJobSeqStepFailed
		step.Error = "canceled"
	}
	skipFrom(run, run.CurrentStep+1, "the run was canceled")
}

func skipFrom(run *entity.SchedJobSeqRun, from int, reason string) {
	for i := from; i < len(run.Steps); i++ {
		run.Steps[i].Status = base.SchedJobSeqStepSkipped
		run.Steps[i].Error = reason
	}
}

// stepLabel is how logs and summaries name a step: its label, else its job.
func stepLabel(step *entity.SchedJobSeqStepResult) string {
	if step.Name != "" {
		return step.Name
	}
	return step.Job.ID
}

// sequenceSummary is one line a step: how each went, for the run's log and its
// notification.
func sequenceSummary(run *entity.SchedJobSeqRun) string {
	lines := make([]string, 0, len(run.Steps))
	for i, step := range run.Steps {
		line := fmt.Sprintf("%d. %s: %s", i+1, stepLabel(step), step.Status)
		switch {
		case step.Error != "":
			line += ": " + step.Error
		case step.Status == base.SchedJobSeqStepDone && !step.EndedAt.IsZero():
			line += fmt.Sprintf(" (%s)", step.EndedAt.Sub(step.StartedAt).Truncate(time.Millisecond))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
