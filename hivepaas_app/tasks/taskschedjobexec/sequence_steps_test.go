package taskschedjobexec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func runOf(statuses ...base.SchedJobSeqStepStatus) *entity.SchedJobSeqRun {
	run := &entity.SchedJobSeqRun{Started: true}
	for i, status := range statuses {
		run.Steps = append(run.Steps, &entity.SchedJobSeqStepResult{
			Job: entity.ObjectID{ID: string(rune('a' + i))}, Status: status,
		})
	}
	return run
}

func TestStepConfigTakesTheMembersRetryAndTimeout(t *testing.T) {
	seqConfig := entity.TaskConfig{Priority: base.TaskPriorityCritical, MaxRetry: 9, Retry: 3,
		Timeout: timeutil.Duration(time.Hour), ControlDisabled: true}
	member := &entity.SchedJob{MaxRetry: 2, RetryDelay: timeutil.Duration(time.Minute),
		RetryDelayIncr: timeutil.Duration(time.Second), Timeout: timeutil.Duration(10 * time.Minute)}

	cfg := stepConfig(seqConfig, member, &entity.SchedJob{Timeout: timeutil.Duration(time.Hour)})

	assert.Equal(t, base.TaskPriorityCritical, cfg.Priority, "the run's priority")
	assert.True(t, cfg.ControlDisabled, "the run's control")
	assert.Equal(t, 2, cfg.MaxRetry)
	assert.Equal(t, 0, cfg.Retry, "a step starts with no retry spent")
	assert.Equal(t, timeutil.Duration(time.Minute), cfg.RetryDelay)
	assert.Equal(t, timeutil.Duration(time.Second), cfg.RetryDelayIncr)
	assert.Equal(t, timeutil.Duration(10*time.Minute), cfg.Timeout)

	cfg = stepConfig(seqConfig, &entity.SchedJob{}, &entity.SchedJob{Timeout: timeutil.Duration(time.Hour)})
	assert.Equal(t, timeutil.Duration(time.Hour), cfg.Timeout, "a member with no timeout: the sequence's")
	assert.Equal(t, 0, cfg.MaxRetry)
}

func TestRetriesLeft(t *testing.T) {
	assert.True(t, retriesLeft(entity.TaskConfig{MaxRetry: 2, Retry: 1}))
	assert.False(t, retriesLeft(entity.TaskConfig{MaxRetry: 2, Retry: 2}))
	assert.False(t, retriesLeft(entity.TaskConfig{}))
}

func TestAfterStepGoesOnWhileStepsAreLeft(t *testing.T) {
	run := runOf(base.SchedJobSeqStepDone, base.SchedJobSeqStepPending)
	assert.Equal(t, seqContinue, afterStep(run, true))
	assert.Equal(t, 1, run.CurrentStep)
}

func TestAfterStepEndsDoneWhenNoStepFailed(t *testing.T) {
	run := runOf(base.SchedJobSeqStepDone, base.SchedJobSeqStepDone)
	run.CurrentStep = 1
	assert.Equal(t, seqDone, afterStep(run, true))
}

func TestAfterStepStopsAtAFailureAndSkipsTheRest(t *testing.T) {
	run := runOf(base.SchedJobSeqStepFailed, base.SchedJobSeqStepPending, base.SchedJobSeqStepPending)

	assert.Equal(t, seqFailed, afterStep(run, true))
	assert.Equal(t, base.SchedJobSeqStepSkipped, run.Steps[1].Status)
	assert.Equal(t, base.SchedJobSeqStepSkipped, run.Steps[2].Status)
	assert.Contains(t, run.Steps[2].Error, "an earlier step")
}

func TestAfterStepCountsASkipAsAFailureWhenStopping(t *testing.T) {
	run := runOf(base.SchedJobSeqStepSkipped, base.SchedJobSeqStepPending)
	assert.Equal(t, seqFailed, afterStep(run, true))
	assert.Equal(t, base.SchedJobSeqStepSkipped, run.Steps[1].Status)
}

func TestAfterStepContinuesPastAFailureAndEndsFailed(t *testing.T) {
	run := runOf(base.SchedJobSeqStepFailed, base.SchedJobSeqStepPending)
	assert.Equal(t, seqContinue, afterStep(run, false))

	run.Steps[1].Status = base.SchedJobSeqStepDone
	assert.Equal(t, seqFailed, afterStep(run, false), "a step failed on the way")
}

func TestAfterStepContinuesPastASkipAndEndsDone(t *testing.T) {
	run := runOf(base.SchedJobSeqStepSkipped, base.SchedJobSeqStepDone)
	assert.Equal(t, seqContinue, afterStep(run, false))
	assert.Equal(t, seqDone, afterStep(run, false), "a skipped step is not a failed one")
}

func TestMarkCanceledEndsTheRun(t *testing.T) {
	run := runOf(base.SchedJobSeqStepDone, base.SchedJobSeqStepRunning, base.SchedJobSeqStepPending)
	run.CurrentStep = 1

	markCanceled(run)

	assert.Equal(t, base.SchedJobSeqStepFailed, run.Steps[1].Status)
	assert.Equal(t, "canceled", run.Steps[1].Error)
	assert.Equal(t, base.SchedJobSeqStepSkipped, run.Steps[2].Status)
	assert.Equal(t, base.SchedJobSeqStepDone, run.Steps[0].Status)
}

func TestSequenceSummary(t *testing.T) {
	start := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	run := runOf(base.SchedJobSeqStepDone, base.SchedJobSeqStepFailed, base.SchedJobSeqStepSkipped)
	run.Steps[0].Name = "migrate"
	run.Steps[0].StartedAt, run.Steps[0].EndedAt = start, start.Add(3*time.Second)
	run.Steps[1].Error = "exit code 2"
	run.Steps[2].Error = "an earlier step failed"

	assert.Equal(t, "1. migrate: done (3s)\n2. b: failed: exit code 2\n3. c: skipped: an earlier step failed",
		sequenceSummary(run))
}

// A step whose job is gone, disabled or itself a sequence is skipped, and says
// why; an active job runs.
func TestMemberSkipReason(t *testing.T) {
	job := func(status base.SettingStatus, kind base.SchedJobType) *entity.Setting {
		return &entity.Setting{Type: base.SettingTypeSchedJob, Status: status, Kind: string(kind)}
	}

	assert.Empty(t, memberSkipReason(job(base.SettingStatusActive, base.SchedJobTypeContainerCommand)))
	assert.Equal(t, "the scheduled job is gone", memberSkipReason(nil))
	assert.Equal(t, "the scheduled job is disabled",
		memberSkipReason(job(base.SettingStatusDisabled, base.SchedJobTypeContainerCommand)))
	assert.Equal(t, "the scheduled job is disabled",
		memberSkipReason(job(base.SettingStatusPending, base.SchedJobTypeContainerCommand)))
	assert.Equal(t, "a sequence cannot run another sequence",
		memberSkipReason(job(base.SettingStatusActive, base.SchedJobTypeJobSequence)))
}
