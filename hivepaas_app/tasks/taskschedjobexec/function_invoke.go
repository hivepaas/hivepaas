package taskschedjobexec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functioninvoke"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/schedjobexecservice"
)

const (
	// invokeBodyMax is how much of a response's body a run keeps, in its task
	// and its log.
	invokeBodyMax = 64 * 1024
	// invokeStepBodyMax is how much of it a step hands on: the steps after it
	// read it in their environment.
	invokeStepBodyMax = 16 * 1024
	// invokeStatusFailed is the first status that fails a run.
	invokeStatusFailed = 400
	// invokeExitBadRequest and invokeExitNotLoaded are call's and invoke's exits
	// without a result: 2, a request that cannot be read - or, from a runtime
	// before 1.2.0, no call; 3, a handler that cannot be loaded (invoke) or a
	// serve that does not answer (call).
	invokeExitBadRequest = 2
	invokeExitNotLoaded  = 3
)

// invokeFunction calls the job's function once: the runtime's call, in a
// running task of the function, reads the job's request on its stdin and hands
// it to the function's serve, whose log - the app's - has the call's lines; on
// a runtime without call, invoke runs the handler itself, its log the run's.
// The response goes to the task's output, or, as a step of a sequence, to the
// steps after it. A status of 400 or more fails the run, as an outcome other
// than ok does.
func (e *Executor) invokeFunction(
	ctx context.Context,
	db database.Tx,
	run *jobRun,
	job *entity.SchedJob,
) (*jobResult, error) {
	if job.FunctionInvoke == nil {
		return &jobResult{}, invokeFailed("the job holds no request to send")
	}
	request, err := json.Marshal(functioninvoke.RequestOf(job.FunctionInvoke))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	logStore := run.execData.LogStore
	// The call goes to the function's serve; a runtime without call - before
	// 1.2.0 - exits 2, as for a request it cannot read, and the request goes to
	// invoke, which refuses that request too.
	output, resp, err := e.execInRuntime(ctx, db, run, job, functioninvoke.CallCommand, request)
	throughServe := true
	if err != nil && exitCodeOf(resp) == invokeExitBadRequest {
		output, resp, err = e.execInRuntime(ctx, db, run, job, functioninvoke.Command, request)
		throughServe = false
	}
	result := &jobResult{}
	if resp != nil {
		result.skipNotification = resp.SkipResultNotification
		result.exitCode = resp.ExitCode
	}
	if err != nil {
		if hint := invokeExitHint(resp, throughServe); hint != "" {
			_ = logStore.Add(ctx, tasklog.NewErrFrame(hint, tasklog.TsNow))
		}
		return result, hperrors.Wrap(err)
	}

	invoked, err := readInvokeResult(output)
	if err != nil {
		_ = logStore.Add(ctx, tasklog.NewErrFrame(err.Error(), tasklog.TsNow))
		return result, hperrors.Wrap(err)
	}
	_ = logStore.Add(ctx, tasklog.NewOutFrame(describeInvokeResult(invoked), tasklog.TsNow))
	if throughServe {
		_ = logStore.Add(ctx, tasklog.NewOutFrame("The call's log is in the app's logs, request "+invoked.RequestID,
			tasklog.TsNow))
	}
	if run.sequence == nil {
		run.execData.Task.MustSetOutput(invoked)
	}
	result.outputs = invokeOutputs(invoked)

	switch {
	case invoked.Outcome != invokeOutcomeOK:
		return result, invokeFailed(fmt.Sprintf("the call ended with outcome %s", invoked.Outcome))
	case invoked.Status >= invokeStatusFailed:
		return result, invokeFailed(fmt.Sprintf("the function answered %d", invoked.Status))
	}
	return result, nil
}

// execInRuntime runs the runtime's command in a running task of the job's
// function, request on its stdin; what it writes is in output, its log lines
// already in the run's log.
func (e *Executor) execInRuntime(
	ctx context.Context,
	db database.Tx,
	run *jobRun,
	job *entity.SchedJob,
	command []string,
	request []byte,
) (*functioninvoke.OutputWriter, *schedjobexecservice.SchedJobExecResp, error) {
	logStore := run.execData.LogStore
	output := functioninvoke.NewOutputWriter(func(line []byte) {
		_ = logStore.Add(ctx, tasklog.NewOutFrame(string(line), tasklog.TsNow))
	})
	resp, err := e.schedJobExecService.SchedJobExec(ctx, db, &schedjobexecservice.SchedJobExecReq{
		TaskExecData:    run.execData,
		SchedJobSetting: run.jobSetting,
		DestApp:         run.refObjects.RefApps[job.App.ID],
		Sequence:        run.sequence,
		Command:         &entity.CommandTemplate{Command: strings.Join(command, " ")},
		Stdin:           bytes.NewReader(request),
		StdoutWriter:    output,
	})
	output.Flush()
	return output, resp, hperrors.Wrap(err)
}

func exitCodeOf(resp *schedjobexecservice.SchedJobExecResp) int {
	if resp == nil || resp.ExitCode == nil {
		return -1
	}
	return *resp.ExitCode
}

// invokeOutcomeOK is the runtime's outcome of a handler that returned.
const invokeOutcomeOK = "ok"

// invokeFailed is a call that failed, saying why.
func invokeFailed(why string) error {
	return hperrors.Wrap(hperrors.ErrInfraActionFailed).WithParam("Error", why)
}

// readInvokeResult is the response invoke wrote, its body cut to what a run
// keeps.
func readInvokeResult(output *functioninvoke.OutputWriter) (*entity.SchedJobFunctionInvokeResult, error) {
	if output.ResultTooLarge() {
		return nil, invokeFailed("the function's response was too large to be read")
	}
	if output.Result() == nil {
		return nil, invokeFailed("the runtime wrote no result of the call")
	}
	parsed, err := functioninvoke.ParseResult(string(output.Result()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	result := &entity.SchedJobFunctionInvokeResult{Outcome: parsed.Outcome, Status: parsed.Status,
		Headers: parsed.Headers, Body: parsed.Body, RequestID: parsed.RequestID, DurationMs: parsed.DurationMs}
	if len(result.Body) > invokeBodyMax {
		result.Body, result.BodyTruncated = result.Body[:invokeBodyMax], true
	}
	return result, nil
}

// describeInvokeResult is the response as the run's log shows it: its status
// and duration, then its body when it is text.
func describeInvokeResult(result *entity.SchedJobFunctionInvokeResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "The function answered %d in %.1f ms (outcome %s)\n", result.Status, result.DurationMs,
		result.Outcome)
	switch {
	case len(result.Body) == 0:
	case utf8.Valid(result.Body):
		sb.Write(result.Body)
		if result.BodyTruncated {
			sb.WriteString("\n… cut at 64 KB")
		}
		sb.WriteString("\n")
	default:
		fmt.Fprintf(&sb, "%d bytes that are not text\n", len(result.Body))
	}
	return sb.String()
}

// invokeOutputs is what a call tells the steps after it: its status, and its
// body when it is text, cut to what an environment holds.
func invokeOutputs(result *entity.SchedJobFunctionInvokeResult) map[string]string {
	outputs := map[string]string{"STATUS": strconv.Itoa(result.Status)}
	if len(result.Body) > 0 && utf8.Valid(result.Body) {
		body := result.Body
		if len(body) > invokeStepBodyMax {
			body = body[:invokeStepBodyMax]
		}
		outputs["BODY"] = string(body)
	}
	return outputs
}

// invokeExitHint says what an exit of invoke without a result means, beside
// the runtime's own message in the log.
func invokeExitHint(resp *schedjobexecservice.SchedJobExecResp, throughServe bool) string {
	switch exitCodeOf(resp) {
	case invokeExitBadRequest:
		return "The runtime could not read the job's request"
	case invokeExitNotLoaded:
		if throughServe {
			return "The function's server does not answer on its instance"
		}
		return "The function's handler could not be loaded: its error is above"
	}
	return ""
}
