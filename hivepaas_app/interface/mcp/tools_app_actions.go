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
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	Effect  string `json:"effect"`
	// Tasks are the app's containers now, as get_app_status answers them.
	Tasks []*appsettingsdto.ServiceTaskResp `json:"tasksNow"`
}

func planRestartAppTool() Tool {
	return planTool("plan_restart_app", "Plan a restart",
		"Plans POST /projects/{project}/{env}/apps/{app}/restart: the swarm replaces the app's containers "+
			"with new ones of the same image and settings, as the dashboard's Restart does - to pick up a "+
			"changed secret or config, or to unstick a container. The plan shows its containers now, as "+
			"get_app_status does. Nothing happens until apply_plan.",
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
				Tasks:  tasks}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/restart"),
				Body:    json.RawMessage(`{}`),
				Summary: fmt.Sprintf("restart %s in %s/%s", ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}

// ---- plan_redeploy_app ----

// redeployDescs describe the deploy endpoint's request.
var redeployDescs = map[string]string{
	"noCache": "a repository app: true to build without the build cache; an image app ignores it",
	"changeId": "the change this deploys, kept on the deployment's trigger. pr-<number> names a pull request " +
		"of the app's repository, which HivePaaS then comments on with how the deployment went",
}

// deploySource is where an app's image comes from, as its deployment settings
// say: what a redeploy's apply checks has not moved.
type deploySource struct {
	Method string `json:"method"`
	Image  string `json:"image,omitempty"`
	Repo   string `json:"repo,omitempty"`
	Ref    string `json:"ref,omitempty"`
	Commit string `json:"commit,omitempty"`
}

type redeployPlan struct {
	App     string `json:"app"`
	Project string `json:"project"`
	Env     string `json:"env"`
	// Request is the deploy endpoint's request the plan sends.
	Request *appactiondto.DeployAppReq `json:"request"`
	// Settings are the app's deployment settings, which the deployment follows,
	// as get_app_settings answers them.
	Settings *appsettingsdto.DeploymentSettingsResp `json:"settings"`
	Effect   string                                 `json:"effect"`
	Issues   []string                               `json:"issues,omitempty"`
}

func planRedeployAppTool() Tool {
	return planToolWith("plan_redeploy_app", "Plan a redeploy",
		"Plans POST /projects/{project}/{env}/apps/{app}/deploy: a new deployment of the app from its "+
			"source as its deployment settings say - an image app pulls its image again, a repository app is "+
			"built from its branch's newest commit - and, once it succeeds, the app's containers replaced. "+
			"The plan shows those settings; to deploy another tag or branch, change them with "+
			"plan_update_app_settings first. Applied, it answers the deployment's id and the task that "+
			"carries it out; a plan whose settings changed since is refused. Nothing happens until apply_plan.",
		NeedExecute, &applier{check: checkDeploySource,
			follow: "get_app_deployment with the deploymentId shows how it goes; get_app_deployment_logs its log."},
		bodyInput(underApp, &appactiondto.DeployAppReq{}, redeployDescs, nil),
		func(ctx context.Context, call *Call, in map[string]any) (redeployPlan, *storedPlan, error) {
			project, _ := in[argProject].(string)
			env, _ := in[argEnv].(string)
			app, _ := in[argApp].(string)
			ref, err := resolveApp(ctx, call, project, env, app)
			if err != nil {
				return redeployPlan{}, nil, err
			}
			body := &appactiondto.DeployAppReq{}
			if err = decodeBody(in, body); err != nil {
				return redeployPlan{}, nil, err
			}
			settings, err := readDeploymentSettings(ctx, call, ref.path(""))
			if err != nil {
				return redeployPlan{}, nil, err
			}
			return makeRedeployPlan(ref, settings, body)
		})
}

func makeRedeployPlan(ref *appRef, settings *appsettingsdto.DeploymentSettingsResp, body *appactiondto.DeployAppReq) (
	redeployPlan, *storedPlan, error) {
	out := redeployPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Request: body, Settings: settings,
		Effect: "a new deployment is made; the app's containers are replaced once it succeeds"}
	source := deploySourceOf(settings)
	if source.Method == "" {
		out.Issues = append(out.Issues, "the app has no source to deploy from yet; set one in its deployment settings")
		return out, nil, nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	seen, err := json.Marshal(redeployCheck{AppPath: ref.path(""), Seen: source})
	if err != nil {
		return redeployPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/deploy"), Body: raw, Check: seen,
		Summary: fmt.Sprintf("redeploy %s in %s/%s from %s", ref.AppKey, ref.ProjectKey, ref.Env,
			source.describe())}, nil
}

// describe is a source in a few words, for a summary.
func (s deploySource) describe() string {
	switch {
	case s.Method == string(base.DeploymentMethodImage):
		return s.Image
	case s.Method == string(base.DeploymentMethodFunction) && s.Repo == "":
		return "the function's code"
	}
	return strings.TrimSpace(s.Repo + " " + s.Ref)
}

// readDeploymentSettings reads an app's deployment settings: what the next
// deployment uses.
func readDeploymentSettings(ctx context.Context, call *Call, appPath string) (
	*appsettingsdto.DeploymentSettingsResp, error) {
	var resp appsettingsdto.GetAppDeploymentSettingsResp
	if err := call.Get(ctx, appPath+"/deployment-settings", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// deploySourceOf is the source deployment settings say, in brief.
func deploySourceOf(s *appsettingsdto.DeploymentSettingsResp) deploySource {
	if s == nil {
		return deploySource{}
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
	case base.DeploymentMethodFunction:
	}
	return out
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
	settings, err := readDeploymentSettings(ctx, call, check.AppPath)
	if err != nil {
		return err
	}
	if deploySourceOf(settings) != check.Seen {
		return planMoved("the app's source")
	}
	return nil
}

// ---- plan_set_app_running ----

type setRunningPlan struct {
	// Request is the endpoint's request the plan sends.
	Request *appactiondto.SetAppRunningReq `json:"request"`
	// App is the app now, as get_app answers it with getStats.
	App    *appdto.AppResp `json:"app"`
	Effect string          `json:"effect"`
}

func planSetAppRunningTool() Tool {
	return planToolWith("plan_set_app_running", "Plan stopping or starting an app",
		"Plans POST /projects/{project}/{env}/apps/{app}/running-status: stopping an app - its containers "+
			"are removed, its settings, volumes and data kept - or starting it again, as the dashboard's "+
			"Stop and Start do. The plan shows the app now, with its running and desired containers (stats). "+
			"Nothing happens until apply_plan.",
		NeedExecute, &applier{follow: "get_app_status shows its containers go or come back."},
		bodyInput(underApp, &appactiondto.SetAppRunningReq{},
			map[string]string{"running": "true to start the app, false to stop it"}, []string{"running"}),
		func(ctx context.Context, call *Call, in map[string]any) (setRunningPlan, *storedPlan, error) {
			project, _ := in[argProject].(string)
			env, _ := in[argEnv].(string)
			app, _ := in[argApp].(string)
			ref, err := resolveApp(ctx, call, project, env, app)
			if err != nil {
				return setRunningPlan{}, nil, err
			}
			body := &appactiondto.SetAppRunningReq{}
			if err = decodeBody(in, body); err != nil {
				return setRunningPlan{}, nil, err
			}
			var resp appdto.GetAppResp
			if err = call.Get(ctx, ref.path(""), url.Values{paramGetStats: {paramTrue}}, &resp); err != nil {
				return setRunningPlan{}, nil, err
			}
			out := setRunningPlan{Request: body, App: resp.Data}
			verb := "stop"
			out.Effect = "the app's containers are removed; its settings, volumes and data are kept"
			if body.Running {
				verb = "start"
				out.Effect = "the app's containers are created again from its current settings"
			}
			raw, err := json.Marshal(body)
			if err != nil {
				return setRunningPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/running-status"), Body: raw,
				Summary: fmt.Sprintf("%s %s in %s/%s", verb, ref.AppKey, ref.ProjectKey, ref.Env)}, nil
		})
}
