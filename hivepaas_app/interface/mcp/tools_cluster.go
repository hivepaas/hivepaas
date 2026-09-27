package mcp

import (
	"context"
	"net/url"
	"strings"
)

// ---- get_task_logs ----

type taskLogsInput struct {
	Task    string `json:"task" jsonschema:"the task's id (data[].id) from list_tasks"`
	Project string `json:"project,omitempty" jsonschema:"with env: read the task under that env, for a key limited to it"`
	Env     string `json:"env,omitempty" jsonschema:"the env's name, with project"`
	App     string `json:"app,omitempty" jsonschema:"with project and env: read the task under that app, for a key limited to it"` //nolint:lll
	logParams
}

func getTaskLogsTool() Tool {
	return readTool("get_task_logs", "Read a task's log",
		"GET /system/tasks/{task}/logs, or the same under /projects/{project}/{env} or "+
			"/projects/{project}/{env}/apps/{app} when they are given - as list_tasks and get_task read the "+
			"task. What a background task printed: a deployment's build, a backup, a scheduled job's command. "+
			"Oldest first, each line with its time and those written to stderr marked [stderr]; a task still "+
			"running answers what it printed so far. "+descLogAnswer,
		func(ctx context.Context, call *Call, in taskLogsInput) (logsAnswer, error) {
			id := strings.TrimSpace(in.Task)
			if id == "" {
				return logsAnswer{}, &InputError{Message: "task is required; list_tasks lists them"}
			}
			q, err := in.query()
			if err != nil {
				return logsAnswer{}, err
			}
			base, err := tasksPath(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return logsAnswer{}, err
			}
			return q.read(ctx, call, base+"/"+url.PathEscape(id)+"/logs")
		})
}

// tasksPath is the task list the names given say: an app's, an env's, or
// every task.
func tasksPath(ctx context.Context, call *Call, project, env, app string) (string, error) {
	project, env, app = strings.TrimSpace(project), strings.TrimSpace(env), strings.TrimSpace(app)
	switch {
	case app != "":
		ref, err := resolveApp(ctx, call, project, env, app)
		if err != nil {
			return "", err
		}
		return ref.path("/tasks"), nil
	case project != "" || env != "":
		ref, err := resolveEnv(ctx, call, project, env)
		if err != nil {
			return "", err
		}
		return ref.path("/tasks"), nil
	}
	return "/system/tasks", nil
}
