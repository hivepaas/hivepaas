package taskschedjobexec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// savedLogs stands for the task log repository: it keeps the lines saved.
type savedLogs struct {
	repository.TaskLogRepo
	lines []string
}

func (s *savedLogs) InsertMulti(
	_ context.Context, _ database.IDB, logs []*entity.TaskLog, _ ...bunex.InsertQueryOption,
) error {
	for _, log := range logs {
		s.lines = append(s.lines, log.Data)
	}
	return nil
}

func loggedRun(goesOn bool) *taskData {
	data := &taskData{
		TaskExecData: &queue.TaskExecData{
			Task:     &entity.Task{ID: "t1", StartedAt: time.Now()},
			LogStore: tasklog.NewLocalStore("task:t1:log"),
		},
		SchedJob: &entity.Setting{ID: "j1"},
	}
	_ = data.LogStore.Add(context.Background(), tasklog.NewOutFrame("step 1 done\n", tasklog.TsNow))
	if goesOn {
		data.Continue()
	}
	return data
}

// A run that goes on - a sequence's next step - keeps its log open: its lines
// are saved, with no end written, and those following it read on. The last
// run writes the end.
func TestARunThatGoesOnKeepsItsLogOpen(t *testing.T) {
	saved := &savedLogs{}
	e := &Executor{taskLogRepo: saved}

	going := loggedRun(true)
	assert.NoError(t, e.saveLogs(context.Background(), nil, going, true))
	assert.Equal(t, []string{"step 1 done\n"}, saved.lines, "no end yet")
	assert.True(t, keepsLogOpen(going))

	saved.lines = nil
	last := loggedRun(false)
	assert.NoError(t, e.saveLogs(context.Background(), nil, last, true))
	assert.Contains(t, saved.lines, "step 1 done\n")
	assert.Regexp(t, "^Job execution finished in", saved.lines[len(saved.lines)-1])
	assert.False(t, keepsLogOpen(last))

	canceled := loggedRun(true)
	canceled.TaskCanceled = true
	assert.False(t, keepsLogOpen(canceled), "a canceled task runs no more")
}
