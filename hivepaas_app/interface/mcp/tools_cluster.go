package mcp

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ---- list_attention ----

type namedObject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// attentionItem is one card of the Home page's attention list, as the API
// answers it.
type attentionItem struct {
	Kind              string       `json:"kind"`
	Severity          string       `json:"severity"`
	Scope             string       `json:"scope"`
	Project           *namedObject `json:"project,omitempty"`
	Env               string       `json:"env,omitempty"`
	App               *namedObject `json:"app,omitempty"`
	Subject           string       `json:"subject"`
	Running           uint64       `json:"running,omitempty"`
	Desired           uint64       `json:"desired,omitempty"`
	Restarts          int          `json:"restarts,omitempty"`
	LastError         string       `json:"lastError,omitempty"`
	NodeState         string       `json:"nodeState,omitempty"`
	MemoryLimitsBytes int64        `json:"memoryLimitsBytes,omitempty"`
	MemoryTotalBytes  int64        `json:"memoryTotalBytes,omitempty"`
	Since             *time.Time   `json:"since,omitempty"`
}

type attentionList struct {
	Items     []attentionItem `json:"items"`
	Truncated int             `json:"truncated,omitempty"`
}

func (l *attentionList) shrink() bool {
	return shrinkList(&l.Items, &l.Truncated)
}

type noInput struct{}

func listAttentionTool() Tool {
	return readTool("list_attention", "What needs attention",
		"Lists what the Home page says needs attention and the API key's user may see: apps whose "+
			"containers do not all run or keep restarting, nodes that are down, memory promised past "+
			"what a node has. Start here when asked what is wrong.",
		func(ctx context.Context, call *Call, _ noInput) (attentionList, error) {
			var resp struct {
				Data struct {
					Items []attentionItem `json:"items"`
				} `json:"data"`
			}
			if err := call.Get(ctx, "/home/attention", nil, &resp); err != nil {
				return attentionList{}, err
			}
			out := attentionList{Items: resp.Data.Items}
			if out.Items == nil {
				out.Items = []attentionItem{}
			}
			for i := range out.Items {
				out.Items[i].LastError = cutText(out.Items[i].LastError, maxErrorText, "")
			}
			return out, nil
		})
}

// ---- list_tasks ----

// maxTaskList is the most tasks list_tasks answers.
const maxTaskList = 50

type listTasksInput struct {
	Project string   `json:"project,omitempty" jsonschema:"a project's key, name or id, with env; empty for every task"`
	Env     string   `json:"env,omitempty" jsonschema:"the env's name, with project"`
	Status  []string `json:"status,omitempty" jsonschema:"not-started, in-progress, done, failed or canceled"`
	Type    []string `json:"type,omitempty" jsonschema:"only these types, such as task:app-deploy or task:sched-job-exec"`
	Limit   int      `json:"limit,omitempty" jsonschema:"how many of the newest to answer, 1-50; 20 when not given"`
}

type taskListItem struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Job       string     `json:"job,omitempty"`
	Project   string     `json:"project,omitempty"`
	App       string     `json:"app,omitempty"`
	LastError string     `json:"lastError,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

type taskList struct {
	Tasks     []taskListItem `json:"tasks"`
	Truncated int            `json:"truncated,omitempty"`
}

func (l *taskList) shrink() bool {
	return shrinkList(&l.Tasks, &l.Truncated)
}

type apiTask struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	TargetJob *struct {
		Name string `json:"name"`
	} `json:"targetJob"`
	LastError    string `json:"lastError"`
	ScopeProject *struct {
		Key string `json:"key"`
	} `json:"scopeProject"`
	ScopeApp *struct {
		Key string `json:"key"`
	} `json:"scopeApp"`
	CreatedAt time.Time  `json:"createdAt"`
	StartedAt *time.Time `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`
}

func listTasksTool() Tool {
	return readTool("list_tasks", "List background tasks",
		"Lists the newest background tasks - deployments, backups, scheduled jobs, cleanups - with "+
			"their state and the error that ended a failed one. Every task, which needs access to "+
			"System, or those of one env. get_task_logs reads what a task printed.",
		func(ctx context.Context, call *Call, in listTasksInput) (taskList, error) {
			limit := in.Limit
			switch {
			case limit == 0:
				limit = 20
			case limit < 0 || limit > maxTaskList:
				return taskList{}, &InputError{Message: "limit is 1 to " + strconv.Itoa(maxTaskList)}
			}
			path, err := tasksPath(ctx, call, in.Project, in.Env)
			if err != nil {
				return taskList{}, err
			}
			query := url.Values{paramPageLimit: {strconv.Itoa(limit)}}
			for _, status := range in.Status {
				query.Add("status", status)
			}
			for _, typ := range in.Type {
				query.Add("type", typ)
			}
			var resp struct {
				Data []apiTask `json:"data"`
			}
			if err = call.Get(ctx, path, query, &resp); err != nil {
				return taskList{}, err
			}
			out := taskList{Tasks: make([]taskListItem, 0, len(resp.Data))}
			for _, t := range resp.Data {
				out.Tasks = append(out.Tasks, makeTaskListItem(&t))
			}
			return out, nil
		})
}

func makeTaskListItem(t *apiTask) taskListItem {
	item := taskListItem{ID: t.ID, Type: t.Type, Status: t.Status, CreatedAt: t.CreatedAt,
		StartedAt: t.StartedAt, EndedAt: t.EndedAt, LastError: cutText(t.LastError, maxErrorText, "")}
	if t.TargetJob != nil {
		item.Job = t.TargetJob.Name
	}
	if t.ScopeProject != nil {
		item.Project = t.ScopeProject.Key
	}
	if t.ScopeApp != nil {
		item.App = t.ScopeApp.Key
	}
	return item
}

// tasksPath is the task list of one env when one is named, or every task.
func tasksPath(ctx context.Context, call *Call, project, env string) (string, error) {
	if project == "" && env == "" {
		return "/system/tasks", nil
	}
	ref, err := resolveEnv(ctx, call, project, env)
	if err != nil {
		return "", err
	}
	return ref.path("/tasks"), nil
}

// ---- get_task_logs ----

type taskLogsInput struct {
	Task    string `json:"task" jsonschema:"the task's id, from list_tasks"`
	Project string `json:"project,omitempty" jsonschema:"the task's project, for a key that reads one env only"`
	Env     string `json:"env,omitempty" jsonschema:"the task's env, with project"`
	Tail    int    `json:"tail,omitempty" jsonschema:"how many of the newest lines to answer, 1-500; 100 when not given"`
	Since   string `json:"since,omitempty" jsonschema:"only lines since then: an RFC 3339 time, or a duration ago like 2h"`
	Grep    string `json:"grep,omitempty" jsonschema:"only lines containing this, ignoring case; /expr/ for a regexp"`
}

func getTaskLogsTool() Tool {
	return readTool("get_task_logs", "Read a task's logs",
		"Reads what a background task printed - a deployment's build, a backup - newest lines last, "+
			"grep applied before the tail as get_app_logs does.",
		func(ctx context.Context, call *Call, in taskLogsInput) (logsAnswer, error) {
			id := strings.TrimSpace(in.Task)
			if id == "" {
				return logsAnswer{}, &InputError{Message: "task is required; list_tasks lists them"}
			}
			q, err := newLogQuery(in.Tail, in.Since, in.Grep)
			if err != nil {
				return logsAnswer{}, err
			}
			base, err := tasksPath(ctx, call, in.Project, in.Env)
			if err != nil {
				return logsAnswer{}, err
			}
			out, err := q.read(ctx, call, base+"/"+url.PathEscape(id)+"/logs")
			out.Task = id
			return out, err
		})
}

// ---- list_nodes ----

type nodeItem struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	Addr         string            `json:"addr"`
	Role         string            `json:"role"`
	Leader       bool              `json:"leader,omitempty"`
	State        string            `json:"state"`
	Availability string            `json:"availability"`
	CPUs         int64             `json:"cpus,omitempty"`
	MemoryBytes  int64             `json:"memoryBytes,omitempty"`
	Platform     string            `json:"platform,omitempty"`
	Engine       string            `json:"engine,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

type nodeList struct {
	Nodes     []nodeItem `json:"nodes"`
	Truncated int        `json:"truncated,omitempty"`
}

func (l *nodeList) shrink() bool {
	return shrinkList(&l.Nodes, &l.Truncated)
}

type apiNode struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Hostname     string            `json:"hostname"`
	Addr         string            `json:"addr"`
	Role         string            `json:"role"`
	IsLeader     bool              `json:"isLeader"`
	State        string            `json:"state"`
	Availability string            `json:"availability"`
	Labels       map[string]string `json:"labels"`
	Platform     *struct {
		Architecture string `json:"architecture"`
		OS           string `json:"os"`
	} `json:"platform"`
	Resources *struct {
		CPUs        int64 `json:"cpus"`
		MemoryBytes int64 `json:"memoryBytes"`
	} `json:"resources"`
	EngineDesc *struct {
		EngineVersion string `json:"engineVersion"`
	} `json:"engineDesc"`
}

func listNodesTool() Tool {
	return readTool("list_nodes", "List cluster nodes",
		"Lists the nodes of the swarm: role, state, whether they take work, their CPUs and memory, "+
			"and their labels, which placement constraints match against.",
		func(ctx context.Context, call *Call, _ noInput) (nodeList, error) {
			var resp struct {
				Data []apiNode `json:"data"`
			}
			query := url.Values{paramPageLimit: {strconv.Itoa(maxListed)}}
			if err := call.Get(ctx, "/cluster/nodes", query, &resp); err != nil {
				return nodeList{}, err
			}
			out := nodeList{Nodes: make([]nodeItem, 0, len(resp.Data))}
			for _, n := range resp.Data {
				out.Nodes = append(out.Nodes, makeNodeItem(&n))
			}
			return out, nil
		})
}

func makeNodeItem(n *apiNode) nodeItem {
	item := nodeItem{ID: n.ID, Name: n.Name, Hostname: n.Hostname, Addr: n.Addr, Role: n.Role,
		Leader: n.IsLeader, State: n.State, Availability: n.Availability, Labels: n.Labels}
	if n.Resources != nil {
		item.CPUs, item.MemoryBytes = n.Resources.CPUs, n.Resources.MemoryBytes
	}
	if n.Platform != nil {
		item.Platform = n.Platform.OS + "/" + n.Platform.Architecture
	}
	if n.EngineDesc != nil {
		item.Engine = n.EngineDesc.EngineVersion
	}
	return item
}
