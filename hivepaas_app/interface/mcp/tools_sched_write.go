package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// schedJobArgs are the arguments of plan_create_sched_job: the app's names,
// and the create endpoint's request body.
type schedJobArgs map[string]any

// descCommand describes a command template, as a job runs one; prefix is its
// path in the request.
func descCommand(prefix string) map[string]string {
	return map[string]string{
		prefix + "command": "a command line, run in the app's container with its shell. Give it or script, " +
			"not both",
		prefix + "script":     "instead of command: a script of several lines, run by the container's shell",
		prefix + "workingDir": "the directory it runs in; the container's own when not given",
		prefix + "envVars": "environment variables it runs with, beside the app's own: each a key, a value " +
			"and isLiteral. A value may refer to the app's variables as ${KEY} and to secrets as " +
			"${secrets.KEY}; isLiteral true takes it as written",
		prefix + "argGroups": "arguments put together into one environment variable: each group, when " +
			"enabled, joins the args it uses - name, or name, separator (a space when not given) and the " +
			"value quoted - with spaces, into the variable exportEnv, which the command can then use",
		prefix + "consoleSize": "the terminal's width and height, for a command that runs with tty",
		prefix + "tty":         "true to run it with a terminal",
		prefix + "link":        "a link about it, such as its documentation",
		prefix + "desc":        "a description of it",
		prefix + argName:       "not used here: set by the endpoint",
		prefix + argKind:       "not used here: set by the endpoint",
	}
}

var schedJobDescs = func() map[string]string {
	descs := map[string]string{
		argName: "the job's name, up to 100 characters",
		"schedule": "when it runs: cronExpr or interval, from initialTime; explain_schedule shows the runs of " +
			"one",
		"schedule.endTime": "when the schedule ends, RFC 3339: no run after it; never when not given",
		"priority": "the priority of its runs in the task queue: " + statusValues(base.AllTaskPriorities) +
			"; default when not given",
		"maxRetry": "how many times a failed run is tried again, 1-100; not retried when not given",
		"retryDelay": "how long a retry waits, up to 24h, such as 30s or 5m; the nth retry waits it, or " +
			"longer as retryDelayIncr or retryBackoff say",
		"retryDelayIncr": "added to the wait for each further retry: the nth waits retryDelay + (n-1) x " +
			"retryDelayIncr. It takes precedence over retryBackoff",
		"retryBackoff": "true to double the wait for each further retry: the nth waits retryDelay x 2^(n-1), " +
			"plus a random part of up to retryBackoffJitter",
		"retryBackoffJitter": "with retryBackoff, the most random time added to a wait; 1s when not given",
		"retryDelayMax":      "the longest a retry waits, however the wait grows",
		"timeout":            "how long a run may take before it is stopped, up to 24h; 3h when not given",
		"controlDisabled":    "true for runs that cannot be canceled once started",
		"command":            "what runs, in the app's container",
		"commandOutput": "what is done with what the command prints, when enabled: saved to a file in a " +
			"storage (saveToFile) or given as input to a command in another app (pipeToApp), one of them",
		"commandOutput.saveToFile": "fileName and filePath in the storage; fileKind, such as postgres-backup, " +
			"for a backup the dashboard can restore; storage, the id of a storage setting and a bucket; " +
			"compressionFormat, one of zstd, gzip, zip, tar or none when not given; encryptionFormat, age or " +
			"none when not given, with its encryptionSecret",
		"commandOutput.pipeToApp": "targetApp, the id of another app the API key's user may change, and " +
			"command, what runs there with the output as its input, given as command is",
		"notification": "who is told how a run went: success and failure, each the id of a notification " +
			"setting; successUseDefault and failureUseDefault, true to use the default notification when no " +
			"id is given. When not given, both use the default, as the dashboard's form starts",
	}
	maps.Copy(descs, scheduleDescs("schedule."))
	maps.Copy(descs, descCommand("command."))
	return descs
}()

// forAudit keeps the values of what may be secret: environment variables, and
// the key a saved output is encrypted with.
func (in schedJobArgs) forAudit() any {
	raw, err := json.Marshal(map[string]any(in))
	if err != nil {
		return nil
	}
	var out map[string]any
	if err = json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	redactEnvVars := func(command any) {
		if c, ok := command.(map[string]any); ok {
			if vars, ok := c["envVars"].([]any); ok {
				for _, v := range vars {
					if env, ok := v.(map[string]any); ok {
						env["value"] = redactedValue
					}
				}
			}
		}
	}
	redactEnvVars(out["command"])
	if output, ok := out["commandOutput"].(map[string]any); ok {
		if file, ok := output["saveToFile"].(map[string]any); ok && file["encryptionSecret"] != nil {
			file["encryptionSecret"] = redactedValue
		}
		if pipe, ok := output["pipeToApp"].(map[string]any); ok {
			redactEnvVars(pipe["command"])
		}
	}
	return out
}

type schedJobPlan struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	// Request is the create endpoint's request the plan sends.
	Request *schedjobdto.SchedJobBaseReq `json:"request"`
	// NextRuns are the job's first runs, as explain_schedule answers them.
	NextRuns []time.Time `json:"nextRuns"`
}

const schedJobRuns = 5

func planCreateSchedJobTool() Tool {
	return planToolWith("plan_create_sched_job", "Plan a scheduled job",
		"Plans POST /projects/{project}/{env}/apps/{app}/sched-jobs: a job that runs a command in the app's "+
			"container on a schedule, as the dashboard's app Scheduled jobs do. The app's scheduled jobs "+
			"feature must be on. The plan is the request - its app, and jobType container-command, are set "+
			"by the tool - and the job's first five runs; initialTime is fixed at the plan's time when not "+
			"given, so the runs shown are the job's. Nothing is created until apply_plan.",
		NeedWrite, &applier{follow: "list_sched_jobs with the app lists it; list_tasks with type " +
			string(base.TaskTypeSchedJobExec) + " and targetId the job's id shows its runs."},
		bodyInput(underApp, &schedjobdto.SchedJobBaseReq{}, schedJobDescs, []string{argName, "schedule", "command"},
			argApp, "jobType"),
		func(ctx context.Context, call *Call, in schedJobArgs) (schedJobPlan, *storedPlan, error) {
			project, _ := in[argProject].(string)
			env, _ := in[argEnv].(string)
			app, _ := in[argApp].(string)
			ref, err := resolveApp(ctx, call, project, env, app)
			if err != nil {
				return schedJobPlan{}, nil, err
			}
			args := maps.Clone(in)
			delete(args, argApp) // the app's name, not the request's app
			body := &schedjobdto.SchedJobBaseReq{}
			if err = decodeBody(args, body); err != nil {
				return schedJobPlan{}, nil, err
			}
			if body.Schedule == nil {
				return schedJobPlan{}, nil, &InputError{Message: "schedule is required"}
			}
			body.App = basedto.ObjectIDReq{ID: ref.AppID}
			body.JobType = base.SchedJobTypeContainerCommand
			if body.Notification == nil { // as the dashboard's form starts
				body.Notification = &basedto.BaseEventNotificationReq{SuccessUseDefault: true, FailureUseDefault: true}
			}
			// The runs a plan shows are the job's only when both start at the
			// same time: the endpoint would start the job when it is created.
			if body.Schedule.InitialTime.IsZero() {
				body.Schedule.InitialTime = timeNow().UTC().Truncate(time.Second)
			}
			var runs schedjobdto.CalcNextRunsResp
			if err = call.Post(ctx, "/settings/sched-jobs/calc-next-runs",
				&schedjobdto.CalcNextRunsReq{ScheduleReq: body.Schedule, Count: schedJobRuns}, &runs); err != nil {
				return schedJobPlan{}, nil, err
			}
			out := schedJobPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Request: body,
				NextRuns: runs.Data}
			raw, err := json.Marshal(&schedjobdto.CreateSchedJobReq{SchedJobBaseReq: body})
			if err != nil {
				return schedJobPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/sched-jobs"), Body: raw,
				Summary: fmt.Sprintf("schedule %q in %s of %s/%s", body.Name, ref.AppKey, ref.ProjectKey,
					ref.Env)}, nil
		})
}
