package queueimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func task(taskType base.TaskType, timeout time.Duration) *entity.Task {
	return &entity.Task{
		Type:   taskType,
		Config: entity.TaskConfig{Timeout: timeutil.Duration(timeout)},
	}
}

// What the task asks for always wins. Anything else would silently shorten a job somebody
// configured on purpose.
func TestResolveTaskTimeout_TaskConfigWins(t *testing.T) {
	// Shorter than the type ceiling.
	assert.Equal(t, 30*time.Second,
		resolveTaskTimeout(task(base.TaskTypeAppDeploy, 30*time.Second)))

	// And longer than it, which is the case that matters: a deploy configured for 2h must not be
	// cut down to the 1h type ceiling.
	assert.Equal(t, 2*time.Hour,
		resolveTaskTimeout(task(base.TaskTypeAppDeploy, 2*time.Hour)))
}

func TestResolveTaskTimeout_FallsBackToTypeCeiling(t *testing.T) {
	assert.Equal(t, timeoutAppDeploy, resolveTaskTimeout(task(base.TaskTypeAppDeploy, 0)))
	assert.Equal(t, timeoutDummy, resolveTaskTimeout(task(base.TaskTypeDummy, 0)))
	assert.Equal(t, timeoutPeriodicExec, resolveTaskTimeout(task(base.TaskTypePeriodicExec, 0)))
}

// Volume clones and user-supplied commands have no duration this code can predict, so they must
// keep the long ceiling rather than start failing.
func TestResolveTaskTimeout_OpenEndedTypesKeepTheLongCeiling(t *testing.T) {
	for _, taskType := range []base.TaskType{
		base.TaskTypeAppClone,
		base.TaskTypeAppPreview,
		base.TaskTypeSchedJobExec,
		base.TaskTypeWorkflow,
	} {
		assert.Equal(t, taskDefaultTimeout, resolveTaskTimeout(task(taskType, 0)),
			"%s must keep the long ceiling", taskType)
	}
}

// An unknown type must still get a bound, never zero - a zero timeout would mean the task holds
// its transaction open forever.
func TestResolveTaskTimeout_UnknownTypeStillBounded(t *testing.T) {
	got := resolveTaskTimeout(task(base.TaskType("task:not-a-real-type"), 0))
	assert.Equal(t, taskDefaultTimeout, got)
	assert.Positive(t, got)
}

// Every declared task type needs an entry, so a new one cannot silently inherit the longest
// possible ceiling by being forgotten here.
func TestTaskTypeTimeouts_CoversEveryDeclaredType(t *testing.T) {
	for _, taskType := range base.AllTaskTypes {
		timeout, ok := taskTypeTimeouts[taskType]
		assert.True(t, ok, "task type %s has no timeout ceiling", taskType)
		assert.Positive(t, timeout, "task type %s has a zero ceiling", taskType)
		assert.LessOrEqual(t, timeout, taskDefaultTimeout,
			"task type %s exceeds the blanket default", taskType)
	}
}
