package schedjobexecserviceimpl

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/commandservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

type commandServiceStub struct{ commandservice.Service }

func (*commandServiceStub) BuildCommand(
	context.Context, database.IDB, *commandservice.BuildCommandReq,
) (*commandservice.BuildCommandResp, error) {
	return &commandservice.BuildCommandResp{}, nil
}

// fakeExec runs nothing: it records what it was asked to run and writes the
// command's "output" to the writer it was given.
type fakeExec struct {
	containerexecservice.Service
	opts   client.ExecCreateOptions
	app    *entity.App
	output string
	// stdin is what the command was given to read.
	stdin string
}

func (f *fakeExec) ContainerExec(
	_ context.Context, req *containerexecservice.ContainerExecReq,
) (*containerexecservice.ContainerExecResp, error) {
	req.ExecOptions(&f.opts)
	f.app = req.App
	if req.StdinReader != nil {
		in, _ := io.ReadAll(req.StdinReader)
		f.stdin = string(in)
	}
	if req.StdoutWriter != nil {
		_, _ = req.StdoutWriter.Write([]byte(f.output))
	}
	return &containerexecservice.ContainerExecResp{ExecStarted: true}, nil
}

// A data backup runs its own command, not the job's, and takes its stdout
// itself, without a TTY.
func TestSchedJobExecRunsAGivenCommandIntoAGivenWriter(t *testing.T) {
	exec := &fakeExec{output: "dump"}
	svc := &service{containerExecService: exec, commandService: &commandServiceStub{}}
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob}
	setting.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeDataBackup, Command: &entity.CommandTemplate{
		Command: "not this", TTY: true,
	}})
	var out bytes.Buffer

	_, err := svc.SchedJobExec(context.Background(), database.Tx{}, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:    &queue.TaskExecData{Task: &entity.Task{ID: "t1"}, LogStore: tasklog.NewNullStore()},
		SchedJobSetting: setting,
		DestApp:         &entity.App{ID: "a1"},
		Command:         &entity.CommandTemplate{Command: "pg_dump app", TTY: true},
		StdoutWriter:    &out,
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"pg_dump", "app"}, exec.opts.Cmd)
	assert.False(t, exec.opts.TTY)
	assert.Equal(t, "dump", out.String())
}
