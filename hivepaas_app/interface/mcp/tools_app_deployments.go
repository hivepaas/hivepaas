package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// The tools of appdeploymentuc's endpoints: an app's deployments, one of them
// with its build's log, and canceling one.

const (
	defaultDeploymentList = 10
	maxDeploymentList     = 50
)

// deploymentItemOf is a deployment in brief, from the endpoint's own type. Its
// settings are read for where it came from and nothing else: they also hold
// commands, which may carry anything.
func deploymentItemOf(d *appdeploymentdto.DeploymentResp) deploymentItem {
	item := deploymentItem{ID: d.ID, Status: string(d.Status), CreatedAt: d.CreatedAt, EndedAt: d.EndedAt}
	if d.Trigger != nil {
		item.Trigger = string(d.Trigger.Source)
	}
	if d.Output != nil {
		item.Error = cutText(d.Output.Error, maxErrorText, "")
		if d.Output.CommitHashShort != "" {
			item.Commit = d.Output.CommitHashShort + " " + d.Output.CommitTitle
		}
	}
	return item
}

// deploymentSourceOf is where a deployment's image came from.
func deploymentSourceOf(d *appdeploymentdto.DeploymentResp) *deploySource {
	s := d.Settings
	if s == nil {
		return nil
	}
	out := &deploySource{Method: string(s.ActiveMethod)}
	switch s.ActiveMethod {
	case base.DeploymentMethodImage:
		if s.ImageSource != nil {
			out.Image = s.ImageSource.Image
		}
	case base.DeploymentMethodRepo:
		if s.RepoSource != nil {
			out.Repo, out.Ref, out.Commit = withoutUserinfo(s.RepoSource.RepoURL), s.RepoSource.RepoRef,
				s.RepoSource.CommitHash
		}
	}
	return out
}

// ---- list_app_deployments ----

type listDeploymentsInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	Limit   int    `json:"limit,omitempty" jsonschema:"how many of the newest, 1-50; 10 when not given"`
}

type deploymentList struct {
	App         string           `json:"app"`
	Deployments []deploymentItem `json:"deployments"`
	Truncated   int              `json:"truncated,omitempty"`
}

func (l *deploymentList) shrink() bool {
	return shrinkList(&l.Deployments, &l.Truncated)
}

func listAppDeploymentsTool() Tool {
	return readTool("list_app_deployments", "List an app's deployments",
		"Lists an app's newest deployments: how each ended, what started it, the commit it built and the "+
			"error that stopped a failed one. get_app_deployment shows one with its build's log.",
		func(ctx context.Context, call *Call, in listDeploymentsInput) (deploymentList, error) {
			limit := in.Limit
			switch {
			case limit == 0:
				limit = defaultDeploymentList
			case limit < 0 || limit > maxDeploymentList:
				return deploymentList{}, &InputError{Message: fmt.Sprintf("limit is 1 to %d", maxDeploymentList)}
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return deploymentList{}, err
			}
			var resp appdeploymentdto.ListDeploymentResp
			if err = call.Get(ctx, ref.path("/deployments"),
				url.Values{paramPageLimit: {strconv.Itoa(limit)}}, &resp); err != nil {
				return deploymentList{}, err
			}
			out := deploymentList{App: ref.AppKey, Deployments: make([]deploymentItem, 0, len(resp.Data))}
			for _, d := range resp.Data {
				if d != nil {
					out.Deployments = append(out.Deployments, deploymentItemOf(d))
				}
			}
			return out, nil
		})
}

// ---- get_app_deployment ----

type deploymentInput struct {
	Project    string `json:"project" jsonschema:"the project's key, name or id"`
	Env        string `json:"env" jsonschema:"the env's name, such as prod"`
	App        string `json:"app" jsonschema:"the app's key, name or id"`
	Deployment string `json:"deployment" jsonschema:"the deployment's id, from list_app_deployments or a redeploy"`
}

type getDeploymentInput struct {
	Project    string `json:"project" jsonschema:"the project's key, name or id"`
	Env        string `json:"env" jsonschema:"the env's name, such as prod"`
	App        string `json:"app" jsonschema:"the app's key, name or id"`
	Deployment string `json:"deployment" jsonschema:"the deployment's id, from list_app_deployments or a redeploy"`
	Tail       int    `json:"tail,omitempty" jsonschema:"how many of the log's last lines, 1-500; 100 when not given"`
	Grep       string `json:"grep,omitempty" jsonschema:"only lines containing this, ignoring case; /expr/ for a regexp"`
}

type deploymentDetail struct {
	App        string         `json:"app"`
	Deployment deploymentItem `json:"deployment"`
	Source     *deploySource  `json:"source,omitempty"`
	Log        logsAnswer     `json:"log"`
}

func getAppDeploymentTool() Tool {
	return readTool("get_app_deployment", "Get a deployment",
		"Shows one of an app's deployments - its state, what it deployed, the error that stopped it - with "+
			"the last lines of its build's log, grep applied before the tail. A deployment still running "+
			"shows how far it got; ask again to follow it.",
		func(ctx context.Context, call *Call, in getDeploymentInput) (deploymentDetail, error) {
			q, err := newLogQuery(in.Tail, "", in.Grep)
			if err != nil {
				return deploymentDetail{}, err
			}
			ref, d, path, err := readDeployment(ctx, call, in.Project, in.Env, in.App, in.Deployment)
			if err != nil {
				return deploymentDetail{}, err
			}
			out := deploymentDetail{App: ref.AppKey, Deployment: deploymentItemOf(d), Source: deploymentSourceOf(d)}
			if out.Log, err = q.read(ctx, call, path+"/logs"); err != nil {
				return deploymentDetail{}, err
			}
			return out, nil
		})
}

// readDeployment finds an app and one of its deployments, and answers the
// deployment's path.
func readDeployment(ctx context.Context, call *Call, project, env, app, id string) (
	*appRef, *appdeploymentdto.DeploymentResp, string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil, "", &InputError{Message: "deployment is required; list_app_deployments lists them"}
	}
	ref, err := resolveApp(ctx, call, project, env, app)
	if err != nil {
		return nil, nil, "", err
	}
	path := ref.path("/deployments/" + url.PathEscape(id))
	var resp appdeploymentdto.GetDeploymentResp
	if err = call.Get(ctx, path, nil, &resp); err != nil {
		return nil, nil, "", err
	}
	if resp.Data == nil {
		return nil, nil, "", &InputError{Message: "no deployment " + id + " of " + ref.AppKey}
	}
	return ref, resp.Data, path, nil
}

// ---- plan_cancel_deployment ----

type cancelDeploymentPlan struct {
	App        string         `json:"app"`
	Deployment deploymentItem `json:"deployment"`
	Effect     string         `json:"effect"`
	Issues     []string       `json:"issues,omitempty"`
}

func planCancelDeploymentTool() Tool {
	return planTool("plan_cancel_deployment", "Plan canceling a deployment",
		"Plans canceling a deployment that has not ended - a build that hangs, a pull that never "+
			"finishes. The app keeps running what it ran before. Nothing happens until apply_plan.",
		NeedWrite, &applier{follow: "get_app_deployment shows it canceled."},
		func(ctx context.Context, call *Call, in deploymentInput) (cancelDeploymentPlan, *storedPlan, error) {
			ref, d, path, err := readDeployment(ctx, call, in.Project, in.Env, in.App, in.Deployment)
			if err != nil {
				return cancelDeploymentPlan{}, nil, err
			}
			out := cancelDeploymentPlan{App: ref.AppKey, Deployment: deploymentItemOf(d),
				Effect: "the deployment stops; the app keeps its current containers"}
			if d.Status != base.DeploymentStatusNotStarted && d.Status != base.DeploymentStatusInProgress {
				out.Issues = append(out.Issues, "the deployment has already ended: "+string(d.Status))
				return out, nil, nil
			}
			return out, &storedPlan{Method: http.MethodPost, Path: path + "/cancel", Body: json.RawMessage(`{}`),
				Summary: fmt.Sprintf("cancel deployment %s of %s in %s/%s", d.ID, ref.AppKey, ref.ProjectKey,
					ref.Env)}, nil
		})
}
