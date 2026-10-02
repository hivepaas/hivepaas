package taskschedjobexec

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// fakeInvoke stands for the job exec service: it takes the command and its
// stdin, and writes what invoke would have written.
type fakeInvoke struct {
	schedjobexecservice.Service
	req      *schedjobexecservice.SchedJobExecReq
	stdin    string
	output   string
	exitCode int
	err      error
}

func (f *fakeInvoke) SchedJobExec(
	_ context.Context, _ database.Tx, req *schedjobexecservice.SchedJobExecReq,
) (*schedjobexecservice.SchedJobExecResp, error) {
	f.req = req
	in, _ := io.ReadAll(req.Stdin)
	f.stdin = string(in)
	_, _ = req.StdoutWriter.Write([]byte(f.output))
	exitCode := f.exitCode
	return &schedjobexecservice.SchedJobExecResp{ExitCode: &exitCode}, f.err
}

// resultLine is the line invoke ends with, for a response.
func resultLine(status int, outcome, body string) string {
	return `#hivepaas-result {"status":` + strconv.Itoa(status) +
		`,"headers":{"content-type":["text/plain"]},"body":"` + base64.StdEncoding.EncodeToString([]byte(body)) +
		`","requestId":"r1","durationMs":4.5,"outcome":"` + outcome + `"}` + "\n"
}

func invokeRun(sequence *schedjobexecservice.SequenceStep) (*jobRun, *entity.Task) {
	task := &entity.Task{ID: "t1", Type: base.TaskTypeSchedJobExec}
	job := &entity.SchedJob{JobType: base.SchedJobTypeFunctionInvoke, App: entity.ObjectID{ID: "fn"},
		FunctionInvoke: &entity.SchedJobFunctionInvoke{Method: "POST", Path: "/report?day=1", Body: "go"}}
	setting := &entity.Setting{ID: "j1", Type: base.SettingTypeSchedJob}
	setting.MustSetData(job)
	refs := entity.NewRefObjects()
	refs.RefApps["fn"] = &entity.App{ID: "fn"}
	return &jobRun{
		execData:   &queue.TaskExecData{Task: task, LogStore: tasklog.NewNullStore()},
		jobSetting: setting,
		refObjects: refs,
		sequence:   sequence,
	}, task
}

func TestAFunctionCallSendsItsRequestAndKeepsTheResponse(t *testing.T) {
	exec := &fakeInvoke{output: "POST /report\n" + resultLine(200, "ok", "done")}
	e := &Executor{schedJobExecService: exec}
	run, task := invokeRun(nil)

	result, err := e.runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	assert.Equal(t, "hivepaas-runtime invoke", exec.req.Command.Command)
	assert.Equal(t, "fn", exec.req.DestApp.ID)
	assert.JSONEq(t, `{"method":"POST","path":"/report","query":{"day":["1"]},"body":"Z28="}`, exec.stdin)
	output, _ := task.OutputAsFunctionInvoke()
	assert.Equal(t, &entity.SchedJobFunctionInvokeResult{Outcome: "ok", Status: 200, Body: []byte("done"),
		Headers: map[string][]string{"content-type": {"text/plain"}}, RequestID: "r1", DurationMs: 4.5}, output)
	assert.Equal(t, map[string]string{"STATUS": "200", "BODY": "done"}, result.outputs)
}

// A status of 400 or more fails the run; the response is kept to be read.
func TestAFunctionCallAnswering400OrMoreFails(t *testing.T) {
	exec := &fakeInvoke{output: resultLine(503, "ok", "down")}
	run, task := invokeRun(nil)

	_, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.Error(t, err)
	output, _ := task.OutputAsFunctionInvoke()
	if assert.NotNil(t, output) {
		assert.Equal(t, 503, output.Status)
	}
}

func TestAFunctionCallThatTimesOutFails(t *testing.T) {
	exec := &fakeInvoke{output: resultLine(500, "timeout", "")}
	run, _ := invokeRun(nil)

	_, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.Error(t, err)
}

// invoke exits 3, with no result, when the handler cannot be loaded.
func TestAFunctionThatCannotBeLoadedFailsTheRun(t *testing.T) {
	exec := &fakeInvoke{exitCode: 3, err: hperrors.Wrap(hperrors.ErrInfraActionFailed)}
	run, task := invokeRun(nil)

	_, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.True(t, errors.Is(err, hperrors.ErrInfraActionFailed), "got %v", err)
	output, _ := task.OutputAsFunctionInvoke()
	assert.Nil(t, output)
}

func TestAFunctionCallWithoutAResultFails(t *testing.T) {
	exec := &fakeInvoke{output: "only a log\n"}
	run, _ := invokeRun(nil)

	_, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.Error(t, err)
}

// As a step of a sequence, the call leaves the task's output - the sequence's
// run - alone, and hands on its status and body.
func TestAFunctionCallStepHandsOnItsResponse(t *testing.T) {
	exec := &fakeInvoke{output: resultLine(200, "ok", `{"rows":3}`)}
	run, task := invokeRun(&schedjobexecservice.SequenceStep{Step: 1, Steps: 2})

	result, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	assert.Empty(t, task.Output)
	assert.Equal(t, map[string]string{"STATUS": "200", "BODY": `{"rows":3}`}, result.outputs)
}

// A large body is kept cut, in the task and in a step's outputs; a body that
// is not text is not handed on.
func TestAFunctionCallsLargeBodyIsCut(t *testing.T) {
	body := strings.Repeat("a", invokeBodyMax+10)
	exec := &fakeInvoke{output: resultLine(200, "ok", body)}
	run, task := invokeRun(nil)

	result, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.NoError(t, err)
	output, _ := task.OutputAsFunctionInvoke()
	assert.Len(t, output.Body, invokeBodyMax)
	assert.True(t, output.BodyTruncated)
	assert.Len(t, result.outputs["BODY"], invokeStepBodyMax)

	exec = &fakeInvoke{output: resultLine(200, "ok", "\xff\xfe")}
	run, _ = invokeRun(nil)
	result, err = (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"STATUS": "200"}, result.outputs)
}

// A job stored without its request - nothing saves one, but a row can be
// written by hand - fails its run without calling the function.
func TestAFunctionCallWithoutItsRequestFails(t *testing.T) {
	exec := &fakeInvoke{output: resultLine(200, "ok", "done")}
	run, _ := invokeRun(nil)
	run.jobSetting.MustSetData(&entity.SchedJob{JobType: base.SchedJobTypeFunctionInvoke,
		App: entity.ObjectID{ID: "fn"}})

	_, err := (&Executor{schedJobExecService: exec}).runJob(context.Background(), database.Tx{}, run)

	assert.True(t, errors.Is(err, hperrors.ErrInfraActionFailed), "got %v", err)
	assert.Nil(t, exec.req, "the function is not called")
}
