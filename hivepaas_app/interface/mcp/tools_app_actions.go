package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appactionuc/appactiondto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// The tools of appactionuc's endpoints: restart, deploy, and set running.

// ---- plan_restart_app ----

type restartPlan struct {
	App     string     `json:"app"`
	Project string     `json:"project"`
	Env     string     `json:"env"`
	Effect  string     `json:"effect"`
	Tasks   []taskItem `json:"tasksNow"`
}

func planRestartAppTool() Tool {
	return planTool("plan_restart_app", "Plan a restart",
		"Plans restarting an app: the swarm replaces its containers with new ones of the same image and "+
			"configuration, as the dashboard's Restart does. Answers its containers now. Nothing happens "+
			"until apply_plan.",
		NeedExecute, &applier{follow: "get_app_status shows the new containers start; get_app_logs what they print."},
		func(ctx context.Context, call *Call, in appInput) (restartPlan, *storedPlan, error) {
			ref, err := in.resolve(ctx, call)
			if err != nil {
				return restartPlan{}, nil, err
			}
			tasks, err := readServiceTasks(ctx, call, ref)
			if err != nil {
				return restartPlan{}, nil, err
			}
			out := restartPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env,
				Effect: "every container of the app is replaced; the app may be unavailable while they start",
				Tasks:  makeAppStatus(ref, tasks).Tasks}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/restart"),
				Body:    json.RawMessage(`{}`),
				Summary: fmt.Sprintf("restart %s in %s/%s", ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}

// ---- plan_redeploy_app ----

type redeployInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	NoCache bool   `json:"noCache,omitempty" jsonschema:"a repository app: build without the build cache"`
}

// deploySource is where an app's image comes from, as its deployment settings
// say: what a redeploy plan shows, and what its apply checks has not moved.
type deploySource struct {
	Method string `json:"method"`
	Image  string `json:"image,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Ref    string `json:"ref,omitempty"`
	Commit string `json:"commit,omitempty"`
}

type redeployPlan struct {
	App     string       `json:"app"`
	Project string       `json:"project"`
	Env     string       `json:"env"`
	Source  deploySource `json:"source"`
	NoCache bool         `json:"noCache,omitempty"`
	Effect  string       `json:"effect"`
	Issues  []string     `json:"issues,omitempty"`
}

func planRedeployAppTool() Tool {
	return planTool("plan_redeploy_app", "Plan a redeploy",
		"Plans deploying an app again from its source as its settings say: an image app pulls its image, "+
			"a repository app is built from its branch. Answers the source. To deploy another tag or branch, "+
			"change the deployment settings with plan_update_app_settings first. Nothing happens until "+
			"apply_plan.",
		NeedExecute, &applier{check: checkDeploySource, result: redeployResult,
			follow: "get_app_deployment with the deploymentId shows how it goes, and its build's log."},
		func(ctx context.Context, call *Call, in redeployInput) (redeployPlan, *storedPlan, error) {
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return redeployPlan{}, nil, err
			}
			source, err := readDeploySource(ctx, call, ref.path(""))
			if err != nil {
				return redeployPlan{}, nil, err
			}
			return makeRedeployPlan(ref, source, in.NoCache)
		})
}

func makeRedeployPlan(ref *appRef, source deploySource, noCache bool) (redeployPlan, *storedPlan, error) {
	out := redeployPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Source: source,
		NoCache: noCache && source.Method == string(base.DeploymentMethodRepo),
		Effect:  "a new deployment is made; the app's containers are replaced once it succeeds"}
	if source.Method == "" {
		out.Issues = append(out.Issues, "the app has no source to deploy from yet; set one in its deployment settings")
		return out, nil, nil
	}
	body, err := json.Marshal(appactiondto.DeployAppReq{NoCache: out.NoCache})
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	seen, err := json.Marshal(redeployCheck{AppPath: ref.path(""), Seen: source})
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/deploy"), Body: body, Check: seen,
		Summary: fmt.Sprintf("redeploy %s in %s/%s from %s", ref.AppKey, ref.ProjectKey, ref.Env,
			source.describe())}, nil
}

// describe is a source in a few words, for a summary.
func (s deploySource) describe() string {
	if s.Method == string(base.DeploymentMethodImage) {
		return s.Image
	}
	return strings.TrimSpace(s.Repo + " " + s.Ref)
}

// readDeploySource reads an app's source from its deployment settings: what the
// next deployment uses.
func readDeploySource(ctx context.Context, call *Call, appPath string) (deploySource, error) {
	var resp appsettingsdto.GetAppDeploymentSettingsResp
	if err := call.Get(ctx, appPath+"/deployment-settings", nil, &resp); err != nil {
		return deploySource{}, err
	}
	s := resp.Data
	if s == nil {
		return deploySource{}, nil
	}
	out := deploySource{Method: string(s.ActiveMethod)}
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
	return out, nil
}

// redeployCheck is what a redeploy plan keeps to check: the source it saw, and
// the app it is of.
type redeployCheck struct {
	AppPath string       `json:"appPath"`
	Seen    deploySource `json:"seen"`
}

// checkDeploySource refuses a redeploy whose app's source changed since the
// plan: the person agreed to deploy what they were shown.
func checkDeploySource(ctx context.Context, call *Call, raw json.RawMessage) error {
	var check redeployCheck
	if err := json.Unmarshal(raw, &check); err != nil {
		return fmt.Errorf("mcp: reading a plan's check: %w", err)
	}
	now, err := readDeploySource(ctx, call, check.AppPath)
	if err != nil {
		return err
	}
	if now != check.Seen {
		return planMoved("the app's source")
	}
	return nil
}

// redeployResult is the deploy endpoint's answer: the deployment, and the task
// that carries it out.
func redeployResult(data json.RawMessage) (any, error) {
	var deployed appactiondto.DeployAppDataResp
	if err := json.Unmarshal(data, &deployed); err != nil {
		return nil, fmt.Errorf("mcp: decoding the answer: %w", err)
	}
	return deployed, nil
}

// ---- plan_set_app_running ----

type setRunningInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	Running bool   `json:"running" jsonschema:"true to start the app, false to stop it"`
}

type setRunningPlan struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	Status  string `json:"statusNow"`
	Running *int   `json:"runningNow,omitempty"`
	Desired *int   `json:"desiredNow,omitempty"`
	Effect  string `json:"effect"`
}

func planSetAppRunningTool() Tool {
	return planTool("plan_set_app_running", "Plan stopping or starting an app",
		"Plans stopping an app - its containers go, its settings and data stay - or starting it again, as "+
			"the dashboard's Stop and Start do. Answers its state now. Nothing happens until apply_plan.",
		NeedExecute, &applier{follow: "get_app_status shows its containers go or come back."},
		func(ctx context.Context, call *Call, in setRunningInput) (setRunningPlan, *storedPlan, error) {
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return setRunningPlan{}, nil, err
			}
			var resp appdto.GetAppResp
			if err = call.Get(ctx, ref.path(""), url.Values{"getStats": {paramTrue}}, &resp); err != nil {
				return setRunningPlan{}, nil, err
			}
			out := setRunningPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env}
			if resp.Data != nil {
				out.Status = string(resp.Data.Status)
				if resp.Data.Stats != nil {
					out.Running, out.Desired = &resp.Data.Stats.RunningTasks, &resp.Data.Stats.DesiredTasks
				}
			}
			verb := "stop"
			out.Effect = "the app's containers are removed; its settings, volumes and data are kept"
			if in.Running {
				verb = "start"
				out.Effect = "the app's containers are created again from its current settings"
			}
			body, err := json.Marshal(appactiondto.SetAppRunningReq{Running: in.Running})
			if err != nil {
				return setRunningPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/running-status"), Body: body,
				Summary: fmt.Sprintf("%s %s in %s/%s", verb, ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}
