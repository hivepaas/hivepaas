package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appdeploymentuc/appdeploymentdto"
)

// The tools of appdeploymentuc's endpoints that are not one GET each: a
// deployment's log, and canceling one. list_app_deployments and
// get_app_deployment are endpoint tools.

type deploymentInput struct {
	Project    string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env        string `json:"env" jsonschema:"the env's name, such as production; list_projects lists each project's envs"` //nolint:lll
	App        string `json:"app" jsonschema:"the app's key, name or id; list_apps lists them"`
	Deployment string `json:"deployment" jsonschema:"the deployment's id, from list_app_deployments or the redeploy that made it"` //nolint:lll
}

// readDeployment finds an app and one of its deployments, and answers the
// deployment's path.
func readDeployment(ctx context.Context, call *Call, in deploymentInput) (
	*appRef, *appdeploymentdto.DeploymentResp, string, error) {
	id := strings.TrimSpace(in.Deployment)
	if id == "" {
		return nil, nil, "", &InputError{Message: "deployment is required; list_app_deployments lists them"}
	}
	ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
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

// ---- get_app_deployment_logs ----

type deploymentLogsInput struct {
	deploymentInput
	logParams
}

func getAppDeploymentLogsTool() Tool {
	return readTool("get_app_deployment_logs", "Read a deployment's log",
		"GET /projects/{project}/{env}/apps/{app}/deployments/{deployment}/logs. What a deployment printed "+
			"as it ran - pulling the image, or cloning and building the repository, then starting the new "+
			"containers - oldest first, each line with its time and those written to stderr marked [stderr]. "+
			"A deployment still running answers what it printed so far; ask again to follow it. "+descLogAnswer,
		func(ctx context.Context, call *Call, in deploymentLogsInput) (logsAnswer, error) {
			q, err := in.query()
			if err != nil {
				return logsAnswer{}, err
			}
			_, _, path, err := readDeployment(ctx, call, in.deploymentInput)
			if err != nil {
				return logsAnswer{}, err
			}
			return q.read(ctx, call, path+"/logs")
		})
}

// ---- plan_cancel_deployment ----

type cancelDeploymentPlan struct {
	// Deployment is the deployment as get_app_deployment answers it.
	Deployment *appdeploymentdto.DeploymentResp `json:"deployment"`
	Effect     string                           `json:"effect"`
	Issues     []string                         `json:"issues,omitempty"`
}

func planCancelDeploymentTool() Tool {
	return planTool("plan_cancel_deployment", "Plan canceling a deployment",
		"Plans POST /projects/{project}/{env}/apps/{app}/deployments/{deployment}/cancel: stopping a "+
			"deployment that has not ended - a build that hangs, an image pull that never finishes. The app "+
			"keeps running what it ran before. The plan shows the deployment as get_app_deployment does; one "+
			"that has already ended ("+string(base.DeploymentStatusDone)+", "+
			string(base.DeploymentStatusFailed)+", "+string(base.DeploymentStatusCanceled)+") cannot be "+
			"canceled, and the plan says so. Nothing happens until apply_plan.",
		NeedWrite, &applier{follow: "get_app_deployment shows it canceled."},
		func(ctx context.Context, call *Call, in deploymentInput) (cancelDeploymentPlan, *storedPlan, error) {
			ref, d, path, err := readDeployment(ctx, call, in)
			if err != nil {
				return cancelDeploymentPlan{}, nil, err
			}
			out := cancelDeploymentPlan{Deployment: d,
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
