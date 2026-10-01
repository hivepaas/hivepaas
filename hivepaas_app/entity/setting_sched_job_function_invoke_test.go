package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A function call's run keeps the response in its task; another job's output is
// not read as one, nor a call's as a snapshot.
func TestAFunctionCallsRunKeepsItsResponse(t *testing.T) {
	result := &SchedJobFunctionInvokeResult{Outcome: "ok", Status: 200, Body: []byte("hi"), DurationMs: 2.5,
		Headers: map[string][]string{"content-type": {"text/plain"}}, RequestID: "r1"}
	task := &Task{}
	task.MustSetOutput(result)

	read, err := (&Task{Output: task.Output}).OutputAsFunctionInvoke()
	assert.NoError(t, err)
	assert.Equal(t, result, read)
	snapshot, err := (&Task{Output: task.Output}).OutputAsDataBackup()
	assert.NoError(t, err)
	assert.Nil(t, snapshot)

	backup := &Task{}
	backup.MustSetOutput(&SchedJobDataBackupResult{SnapshotID: "k1"})
	read, err = (&Task{Output: backup.Output}).OutputAsFunctionInvoke()
	assert.NoError(t, err)
	assert.Nil(t, read)

	read, err = (&Task{}).OutputAsFunctionInvoke()
	assert.NoError(t, err)
	assert.Nil(t, read)
}
