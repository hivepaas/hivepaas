package mcp

import (
	"context"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"time"
)

type envInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
}

type appInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
}

func (in appInput) resolve(ctx context.Context, call *Call) (*appRef, error) {
	return resolveApp(ctx, call, in.Project, in.Env, in.App)
}

// ---- list_apps ----

type appItem struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Engine  string   `json:"engine,omitempty"`
	Status  string   `json:"status"`
	Owner   string   `json:"owner,omitempty"`
	Running *int     `json:"running,omitempty"`
	Desired *int     `json:"desired,omitempty"`
	Links   []string `json:"links,omitempty"`
}

type appList struct {
	Project   string    `json:"project"`
	Env       string    `json:"env"`
	Apps      []appItem `json:"apps"`
	Truncated int       `json:"truncated,omitempty"`
}

func (l *appList) shrink() bool {
	return shrinkList(&l.Apps, &l.Truncated)
}

func listAppsTool() Tool {
	return readTool("list_apps", "List apps",
		"Lists the apps of one env of a project: what each runs, its status, and how many of its "+
			"containers run against how many it should. An app made by another - a template's "+
			"component, a preview - names that app as its owner.",
		func(ctx context.Context, call *Call, in envInput) (appList, error) {
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return appList{}, err
			}
			apps, err := listApps(ctx, call, ref, true)
			if err != nil {
				return appList{}, err
			}
			out := appList{Project: ref.ProjectKey, Env: ref.Env, Apps: make([]appItem, 0, len(apps))}
			for _, a := range apps {
				item := appItem{Key: a.Key, Name: a.Name, Engine: a.Engine, Status: a.Status,
					Owner: a.owner, Links: a.AccessLinks}
				if a.Stats != nil {
					item.Running, item.Desired = &a.Stats.RunningTasks, &a.Stats.DesiredTasks
				}
				out.Apps = append(out.Apps, item)
			}
			return out, nil
		})
}

// ---- get_app ----

// maxDeployments is how many of an app's deployments get_app shows.
const maxDeployments = 5

// maxErrorText is the most of one error message an answer quotes.
const maxErrorText = 2000

type appSource struct {
	Method string `json:"method"`
	Image  string `json:"image,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Ref    string `json:"ref,omitempty"`
}

type deploymentItem struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Trigger   string     `json:"trigger,omitempty"`
	Commit    string     `json:"commit,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type appDetail struct {
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	ID          string           `json:"id"`
	Project     string           `json:"project"`
	Env         string           `json:"env"`
	Engine      string           `json:"engine,omitempty"`
	Status      string           `json:"status"`
	Note        string           `json:"note,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Owner       string           `json:"owner,omitempty"`
	Children    []string         `json:"children,omitempty"`
	Links       []string         `json:"links,omitempty"`
	Running     *int             `json:"running,omitempty"`
	Desired     *int             `json:"desired,omitempty"`
	Source      *appSource       `json:"source,omitempty"`
	Deployments []deploymentItem `json:"deployments"`
}

// apiAppDetail is the app endpoint's answer, as far as get_app reads it.
type apiAppDetail struct {
	apiApp
	ParentApp *struct {
		Key string `json:"key"`
	} `json:"parentApp"`
}

// apiDeployment is a deployment as the API answers it. Its settings are read
// for where the app comes from and nothing else: they also hold commands, which
// may carry anything.
type apiDeployment struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Settings *struct {
		ActiveMethod string `json:"activeMethod"`
		ImageSource  *struct {
			Image string `json:"image"`
		} `json:"imageSource"`
		RepoSource *struct {
			RepoURL string `json:"repoURL"`
			RepoRef string `json:"repoRef"`
		} `json:"repoSource"`
	} `json:"settings"`
	Trigger *struct {
		Source string `json:"source"`
	} `json:"trigger"`
	Output *struct {
		Error           string `json:"error"`
		CommitHashShort string `json:"commitHashShort"`
		CommitTitle     string `json:"commitTitle"`
	} `json:"output"`
	CreatedAt time.Time  `json:"createdAt"`
	EndedAt   *time.Time `json:"endedAt"`
}

func getAppTool() Tool {
	return readTool("get_app", "Get an app",
		"Shows one app: what it runs and where from, its status, its links, the apps it owns, and "+
			"its last five deployments with how each ended. Use get_app_status for why its "+
			"containers do not run, and get_app_logs for what they print.",
		func(ctx context.Context, call *Call, in appInput) (appDetail, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return appDetail{}, err
			}
			var appResp struct {
				Data apiAppDetail `json:"data"`
			}
			if err = call.Get(ctx, ref.path(""), url.Values{"getStats": {paramTrue}}, &appResp); err != nil {
				return appDetail{}, err
			}
			var deployResp struct {
				Data []apiDeployment `json:"data"`
			}
			if err = call.Get(ctx, ref.path("/deployments"),
				url.Values{paramPageLimit: {strconv.Itoa(maxDeployments)}}, &deployResp); err != nil {
				return appDetail{}, err
			}
			return makeAppDetail(ref, &appResp.Data, deployResp.Data), nil
		})
}

func makeAppDetail(ref *appRef, a *apiAppDetail, deployments []apiDeployment) appDetail {
	out := appDetail{Key: a.Key, Name: a.Name, ID: a.ID, Project: ref.ProjectKey, Env: ref.Env,
		Engine: a.Engine, Status: a.Status, Note: a.Note, Tags: a.Tags, Links: a.AccessLinks,
		Deployments: make([]deploymentItem, 0, len(deployments))}
	if a.ParentApp != nil {
		out.Owner = a.ParentApp.Key
	}
	for _, child := range slices.Concat(a.ChildApps, a.LogicalChildApps) {
		out.Children = append(out.Children, child.Key)
	}
	if a.Stats != nil {
		out.Running, out.Desired = &a.Stats.RunningTasks, &a.Stats.DesiredTasks
	}
	for i, d := range deployments {
		if i == 0 && d.Settings != nil {
			out.Source = &appSource{Method: d.Settings.ActiveMethod}
			if d.Settings.ImageSource != nil {
				out.Source.Image = d.Settings.ImageSource.Image
			}
			if d.Settings.RepoSource != nil {
				out.Source.Repo = withoutUserinfo(d.Settings.RepoSource.RepoURL)
				out.Source.Ref = d.Settings.RepoSource.RepoRef
			}
		}
		item := deploymentItem{ID: d.ID, Status: d.Status, CreatedAt: d.CreatedAt, EndedAt: d.EndedAt}
		if d.Trigger != nil {
			item.Trigger = d.Trigger.Source
		}
		if d.Output != nil {
			item.Error = cutText(d.Output.Error, maxErrorText, "")
			if d.Output.CommitHashShort != "" {
				item.Commit = d.Output.CommitHashShort + " " + d.Output.CommitTitle
			}
		}
		out.Deployments = append(out.Deployments, item)
	}
	return out
}

// ---- get_app_status ----

// maxTasks is how many of an app's containers get_app_status shows.
const maxTasks = 20

type taskItem struct {
	ID      string    `json:"id"`
	Slot    int       `json:"slot,omitempty"`
	Node    string    `json:"node,omitempty"`
	State   string    `json:"state"`
	Desired string    `json:"desired"`
	Message string    `json:"message,omitempty"`
	Error   string    `json:"error,omitempty"`
	At      time.Time `json:"at"`
}

type appStatus struct {
	App       string     `json:"app"`
	Tasks     []taskItem `json:"tasks"`
	Truncated int        `json:"truncated,omitempty"`
}

func (s *appStatus) shrink() bool {
	return shrinkList(&s.Tasks, &s.Truncated)
}

type apiServiceTask struct {
	ID   string `json:"id"`
	Slot int    `json:"slot"`
	Node *struct {
		Hostname string `json:"hostname"`
	} `json:"node"`
	Status *struct {
		Timestamp time.Time `json:"timestamp"`
		State     string    `json:"state"`
		Message   string    `json:"message"`
		Err       string    `json:"err"`
	} `json:"status"`
	DesiredState string `json:"desiredState"`
}

func getAppStatusTool() Tool {
	return readTool("get_app_status", "Get an app's containers",
		"Shows the swarm tasks of an app, newest first: which node each is on, the state it is in "+
			"against the state it should be in, and the error that stopped it. This is where a "+
			"container that never starts - an image that cannot be pulled, a port taken, a "+
			"constraint no node meets - says why.",
		func(ctx context.Context, call *Call, in appInput) (appStatus, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return appStatus{}, err
			}
			var resp struct {
				Data []apiServiceTask `json:"data"`
			}
			if err = call.Get(ctx, ref.path("/service-tasks"), nil, &resp); err != nil {
				return appStatus{}, err
			}
			return makeAppStatus(ref, resp.Data), nil
		})
}

func makeAppStatus(ref *appRef, tasks []apiServiceTask) appStatus {
	out := appStatus{App: ref.AppKey, Tasks: make([]taskItem, 0, len(tasks))}
	for _, t := range tasks {
		item := taskItem{ID: shortID(t.ID), Slot: t.Slot, Desired: t.DesiredState}
		if t.Node != nil {
			item.Node = t.Node.Hostname
		}
		if t.Status != nil {
			item.State, item.At = t.Status.State, t.Status.Timestamp
			item.Message = t.Status.Message
			item.Error = cutText(t.Status.Err, maxErrorText, "")
		}
		out.Tasks = append(out.Tasks, item)
	}
	sort.SliceStable(out.Tasks, func(i, j int) bool { return out.Tasks[i].At.After(out.Tasks[j].At) })
	if len(out.Tasks) > maxTasks {
		out.Truncated = len(out.Tasks) - maxTasks
		out.Tasks = out.Tasks[:maxTasks]
	}
	return out
}

// withoutUserinfo is a repository URL without the credentials one may carry,
// as https://user:token@host/repo does.
func withoutUserinfo(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	u.User = nil
	return u.String()
}

// shortID is a Docker id the way docker's own commands print it.
func shortID(id string) string {
	const short = 12
	if len(id) > short {
		return id[:short]
	}
	return id
}
