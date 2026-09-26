package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
)

type createSchedJobInput struct {
	Project  string `json:"project" jsonschema:"the project's key, name or id"`
	Env      string `json:"env" jsonschema:"the env's name, such as prod"`
	App      string `json:"app" jsonschema:"the app whose container runs the command"`
	Name     string `json:"name" jsonschema:"the job's name"`
	CronExpr string `json:"cronExpr,omitempty" jsonschema:"minute hour day month weekday, or @daily and the like"`
	Interval string `json:"interval,omitempty" jsonschema:"instead of cronExpr: a duration such as 90m or 24h"`
	TimeZone string `json:"timeZone,omitempty" jsonschema:"the IANA zone the schedule is read in; UTC when empty"`
	Command  string `json:"command" jsonschema:"the command, run in the app's container"`
	Timeout  string `json:"timeout,omitempty" jsonschema:"how long a run may take, such as 30m; no limit when empty"`
	MaxRetry int    `json:"maxRetry,omitempty" jsonschema:"how many times a failed run is tried again, 0-10"`
}

type schedJobPlan struct {
	App      string   `json:"app"`
	Project  string   `json:"project"`
	Env      string   `json:"env"`
	Name     string   `json:"name"`
	Schedule string   `json:"schedule"`
	TimeZone string   `json:"timeZone"`
	NextRuns []string `json:"nextRuns"`
	Command  string   `json:"command"`
	Timeout  string   `json:"timeout,omitempty"`
	MaxRetry int      `json:"maxRetry,omitempty"`
}

const (
	schedJobRuns     = 5
	maxSchedJobRetry = 10
)

func planCreateSchedJobTool() Tool {
	return planTool("plan_create_sched_job", "Plan a scheduled job",
		"Plans a job that runs a command in an app's container on a schedule: a cron expression, read in "+
			"the time zone given, or an interval. Answers the job and its next five runs. Nothing is "+
			"created until apply_plan.",
		NeedWrite, &applier{follow: "list_sched_jobs lists it; list_tasks with type task:sched-job-exec shows its runs."},
		func(ctx context.Context, call *Call, in createSchedJobInput) (schedJobPlan, *storedPlan, error) {
			if err := in.check(); err != nil {
				return schedJobPlan{}, nil, err
			}
			zone := strings.TrimSpace(in.TimeZone)
			if zone == "" {
				zone = "UTC"
			}
			loc, err := time.LoadLocation(zone)
			if err != nil {
				return schedJobPlan{}, nil, &InputError{Message: fmt.Sprintf("no time zone %q; use an IANA name", zone)}
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return schedJobPlan{}, nil, err
			}
			// A job reads its cron expression in the zone of its initial time:
			// the plan's runs and the job's are computed from the same one.
			sched, err := scheduleReq(strings.TrimSpace(in.CronExpr), strings.TrimSpace(in.Interval),
				timeNow().In(loc).Truncate(time.Second))
			if err != nil {
				return schedJobPlan{}, nil, err
			}
			runs, err := calcNextRuns(ctx, call, sched, schedJobRuns)
			if err != nil {
				return schedJobPlan{}, nil, err
			}

			out := schedJobPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Name: strings.TrimSpace(in.Name),
				Schedule: gofn.Coalesce(strings.TrimSpace(in.CronExpr), "every "+strings.TrimSpace(in.Interval)),
				TimeZone: zone, Command: in.Command, Timeout: in.Timeout, MaxRetry: in.MaxRetry,
				NextRuns: make([]string, 0, len(runs))}
			for _, run := range runs {
				out.NextRuns = append(out.NextRuns, run.In(loc).Format("2006-01-02T15:04:05Z07:00 Mon"))
			}

			// The create endpoint's own request, as the dashboard's form fills it
			// for a command run in the app's container.
			timeout, _ := timeutil.ParseDurationWithEmptyIsZero(strings.TrimSpace(in.Timeout))
			body := &schedjobdto.CreateSchedJobReq{SchedJobBaseReq: &schedjobdto.SchedJobBaseReq{
				Name: out.Name, JobType: base.SchedJobTypeContainerCommand, Schedule: sched,
				App: basedto.ObjectIDReq{ID: ref.AppID}, MaxRetry: in.MaxRetry, Timeout: timeout,
				Command:      &commandtemplatedto.CommandTemplateBaseReq{Command: in.Command},
				Notification: &basedto.BaseEventNotificationReq{SuccessUseDefault: true, FailureUseDefault: true},
			}}
			raw, err := json.Marshal(body)
			if err != nil {
				return schedJobPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/sched-jobs"), Body: raw,
				Summary: fmt.Sprintf("schedule %q in %s of %s/%s, %s", out.Name, ref.AppKey, ref.ProjectKey,
					ref.Env, out.Schedule)}, nil
		})
}

func (in *createSchedJobInput) check() error {
	switch {
	case strings.TrimSpace(in.Name) == "":
		return &InputError{Message: "name is required"}
	case strings.TrimSpace(in.Command) == "":
		return &InputError{Message: "command is required"}
	case (strings.TrimSpace(in.CronExpr) == "") == (strings.TrimSpace(in.Interval) == ""):
		return &InputError{Message: "give either cronExpr or interval"}
	case in.MaxRetry < 0 || in.MaxRetry > maxSchedJobRetry:
		return &InputError{Message: fmt.Sprintf("maxRetry is 0 to %d", maxSchedJobRetry)}
	}
	if t := strings.TrimSpace(in.Timeout); t != "" {
		if d, err := timeutil.ParseDuration(t); err != nil || d <= 0 {
			return &InputError{Message: fmt.Sprintf("timeout is %q; give a duration such as 30m", t)}
		}
	}
	return nil
}
