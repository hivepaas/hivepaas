package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

// ---- plan_run_sched_job ----

type runSchedJobInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string `json:"env" jsonschema:"the env's name, such as production"`
	App     string `json:"app,omitempty" jsonschema:"the job's app, by key, name or id; left out, a job of the env"`
	Job     string `json:"job" jsonschema:"the job's name or id; list_sched_jobs lists an app's"`
}

type runSchedJobPlan struct {
	Project string `json:"project"`
	Env     string `json:"env"`
	App     string `json:"app,omitempty"`
	// Job is the job as the list answers it: its type, schedule, command.
	Job    *schedjobdto.SchedJobResp `json:"job"`
	Effect string                    `json:"effect"`
}

func planRunSchedJobTool() Tool {
	return planTool("plan_run_sched_job", "Plan running a scheduled job now",
		"Plans POST /projects/{project}/{env}[/apps/{app}]/sched-jobs/{job}/exec: running a scheduled job "+
			"once, now, outside its schedule - as the dashboard's Run Now does - such as a backup before a "+
			"change, or a migration job. Its schedule is not changed. The plan shows the job: its type, "+
			"schedule and command. Applied, the endpoint answers the task that runs it. Nothing runs until "+
			"apply_plan.",
		NeedExecute, &applier{follow: "get_task with the task's id shows how the run goes; get_task_logs " +
			"what it printed."},
		func(ctx context.Context, call *Call, in runSchedJobInput) (runSchedJobPlan, *storedPlan, error) {
			env, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return runSchedJobPlan{}, nil, err
			}
			base := env.path("")
			out := runSchedJobPlan{Project: env.ProjectKey, Env: env.Env}
			where := env.ProjectKey + "/" + env.Env
			if strings.TrimSpace(in.App) != "" {
				app, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
				if err != nil {
					return runSchedJobPlan{}, nil, err
				}
				base, out.App, where = app.path(""), app.AppKey, app.AppKey+" in "+where
			}

			var list schedjobdto.ListSchedJobResp
			query := url.Values{paramPageLimit: {strconv.Itoa(maxListed)}}
			if err = call.Get(ctx, base+"/sched-jobs", query, &list); err != nil {
				return runSchedJobPlan{}, nil, err
			}
			job, err := pick("job", in.Job, "list_sched_jobs", slices.DeleteFunc(list.Data, isNil),
				func(j *schedjobdto.SchedJobResp) named {
					if j.BaseSettingResp == nil {
						return named{}
					}
					return named{id: j.ID, key: j.ID, name: j.Name}
				})
			if err != nil {
				return runSchedJobPlan{}, nil, err
			}
			out.Job = job
			out.Effect = "the job runs once now, as a task; its schedule and next runs are unchanged"

			return out, &storedPlan{Method: http.MethodPost,
				Path:    base + "/sched-jobs/" + url.PathEscape(job.ID) + "/exec",
				Body:    json.RawMessage(`{}`),
				Summary: fmt.Sprintf("run the scheduled job %s of %s now", job.Name, where)}, nil
		})
}
