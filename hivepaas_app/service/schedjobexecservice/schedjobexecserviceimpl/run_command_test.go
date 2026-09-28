package schedjobexecserviceimpl

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// A command outside any job - a restore's - runs in the app it is given, reads
// what it is given on its stdin, and has no TTY to mangle it.
func TestRunCommandGivesTheCommandItsStdin(t *testing.T) {
	exec := &fakeExec{}
	svc := &service{containerExecService: exec, commandService: &commandServiceStub{}}
	app := &entity.App{ID: "a2"}

	resp, err := svc.RunCommand(context.Background(), database.Tx{}, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		App:          app,
		Command:      &entity.CommandTemplate{Command: "psql -U app app", TTY: true, WorkingDir: "/tmp"},
		Stdin:        strings.NewReader("CREATE TABLE t();\n"),
	})

	assert.NoError(t, err)
	assert.Same(t, app, exec.app)
	assert.Equal(t, []string{"psql", "-U", "app", "app"}, exec.opts.Cmd)
	assert.Equal(t, "/tmp", exec.opts.WorkingDir)
	assert.False(t, exec.opts.TTY)
	assert.True(t, exec.opts.AttachStdout)
	assert.Equal(t, "CREATE TABLE t();\n", exec.stdin)
	if assert.NotNil(t, resp.ExitCode) {
		assert.Equal(t, 0, *resp.ExitCode)
	}
}

func TestRunCommandRefusesAnEmptyCommand(t *testing.T) {
	svc := &service{containerExecService: &fakeExec{}, commandService: &commandServiceStub{}}

	_, err := svc.RunCommand(context.Background(), database.Tx{}, &schedjobexecservice.RunCommandReq{
		TaskExecData: &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		App:          &entity.App{ID: "a2"},
		Command:      &entity.CommandTemplate{},
	})

	assert.Error(t, err)
}
