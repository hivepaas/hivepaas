package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ---- list_sched_jobs ----

type schedJobsInput struct {
	Project string `json:"project,omitempty" jsonschema:"with env and app, that app's jobs; empty for the global jobs"`
	Env     string `json:"env,omitempty" jsonschema:"the app's env"`
	App     string `json:"app,omitempty" jsonschema:"the app's key, name or id"`
}

type schedule struct {
	CronExpr    string    `json:"cronExpr,omitempty"`
	Interval    string    `json:"interval,omitempty"`
	InitialTime time.Time `json:"initialTime"`
	EndTime     time.Time `json:"endTime,omitzero"`
}

type schedJobItem struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Status   string      `json:"status"`
	JobType  string      `json:"jobType"`
	App      string      `json:"app,omitempty"`
	Schedule *schedule   `json:"schedule,omitempty"`
	NextRuns []time.Time `json:"nextRuns,omitempty"`
	Timeout  string      `json:"timeout,omitempty"`
	MaxRetry int         `json:"maxRetry,omitempty"`
}

type schedJobList struct {
	Jobs      []schedJobItem `json:"jobs"`
	Truncated int            `json:"truncated,omitempty"`
}

func (l *schedJobList) shrink() bool {
	return shrinkList(&l.Jobs, &l.Truncated)
}

// apiSchedJob is a scheduled job as the API answers it. Its command is not
// read: a command may carry anything.
type apiSchedJob struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	JobType  string    `json:"jobType"`
	Schedule *schedule `json:"schedule"`
	App      *struct {
		Name string `json:"name"`
	} `json:"app"`
	NextRuns []time.Time `json:"nextRuns"`
	Timeout  string      `json:"timeout"`
	MaxRetry int         `json:"maxRetry"`
}

func listSchedJobsTool() Tool {
	return readTool("list_sched_jobs", "List scheduled jobs",
		"Lists scheduled jobs - the global ones, or one app's - with their schedules and state. "+
			"list_tasks with type task:sched-job-exec shows how their runs went.",
		func(ctx context.Context, call *Call, in schedJobsInput) (schedJobList, error) {
			path := "/settings/sched-jobs"
			if in.Project != "" || in.Env != "" || in.App != "" {
				ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
				if err != nil {
					return schedJobList{}, err
				}
				path = ref.path("/sched-jobs")
			}
			var resp struct {
				Data []apiSchedJob `json:"data"`
			}
			if err := call.Get(ctx, path, url.Values{paramPageLimit: {strconv.Itoa(maxListed)}}, &resp); err != nil {
				return schedJobList{}, err
			}
			out := schedJobList{Jobs: make([]schedJobItem, 0, len(resp.Data))}
			for _, j := range resp.Data {
				item := schedJobItem{ID: j.ID, Name: j.Name, Status: j.Status, JobType: j.JobType,
					Schedule: j.Schedule, NextRuns: j.NextRuns, Timeout: j.Timeout, MaxRetry: j.MaxRetry}
				if j.App != nil {
					item.App = j.App.Name
				}
				out.Jobs = append(out.Jobs, item)
			}
			return out, nil
		})
}

// ---- explain_schedule ----

const (
	defaultRunCount = 5
	maxRunCount     = 10
)

type explainScheduleInput struct {
	CronExpr string `json:"cronExpr,omitempty" jsonschema:"minute hour day month weekday, or @daily and the like"`
	Interval string `json:"interval,omitempty" jsonschema:"instead of cronExpr: a duration such as 90m or 24h"`
	TimeZone string `json:"timeZone,omitempty" jsonschema:"an IANA zone such as Asia/Ho_Chi_Minh; UTC when empty"`
	Count    int    `json:"count,omitempty" jsonschema:"how many runs, 1-10; 5 when not given"`
}

type scheduleRuns struct {
	TimeZone string   `json:"timeZone"`
	Runs     []string `json:"runs"`
}

func explainScheduleTool() Tool {
	return readTool("explain_schedule", "Explain a schedule",
		"Answers the next runs of a cron expression or an interval, as HivePaaS computes them for a "+
			"scheduled job. A job reads its cron expression in the time zone of its initial time, so "+
			"give the zone the job is set up in.",
		func(ctx context.Context, call *Call, in explainScheduleInput) (scheduleRuns, error) {
			count := in.Count
			switch {
			case count == 0:
				count = defaultRunCount
			case count < 0 || count > maxRunCount:
				return scheduleRuns{}, &InputError{Message: fmt.Sprintf("count is 1 to %d", maxRunCount)}
			}
			cronExpr, interval := strings.TrimSpace(in.CronExpr), strings.TrimSpace(in.Interval)
			if (cronExpr == "") == (interval == "") {
				return scheduleRuns{}, &InputError{Message: "give either cronExpr or interval"}
			}
			zone := strings.TrimSpace(in.TimeZone)
			if zone == "" {
				zone = "UTC"
			}
			loc, err := time.LoadLocation(zone)
			if err != nil {
				return scheduleRuns{}, &InputError{Message: fmt.Sprintf("no time zone %q; use an IANA name", zone)}
			}
			body := map[string]any{"count": count, "initialTime": timeNow().In(loc).Format(time.RFC3339)}
			if cronExpr != "" {
				body["cronExpr"] = cronExpr
			} else {
				body["interval"] = interval
			}
			var resp struct {
				Data []time.Time `json:"data"`
			}
			if err = call.Post(ctx, "/settings/sched-jobs/calc-next-runs", body, &resp); err != nil {
				return scheduleRuns{}, err
			}
			out := scheduleRuns{TimeZone: zone, Runs: make([]string, 0, len(resp.Data))}
			for _, run := range resp.Data {
				out.Runs = append(out.Runs, run.In(loc).Format("2006-01-02T15:04:05Z07:00 Mon"))
			}
			return out, nil
		})
}
