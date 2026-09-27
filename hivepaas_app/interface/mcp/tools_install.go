package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

type installPlan struct {
	Project string `json:"project"`
	Env     string `json:"env"`
	// Request is the create endpoint's request the plan sends.
	Request *apptemplatedto.CreateAppFromTemplateReq `json:"request"`
	// Preflight is what preflight_install answers for the request.
	Preflight *apptemplatedto.PreflightAppResult `json:"preflight"`
	Warnings  []string                           `json:"warnings,omitempty"`
}

const leftoverDataWarning = "Some of these apps would find data an earlier install left in their " +
	"directories (preflight.storage), and start with it: a database keeps its old data and the password it " +
	"was created with, not the one generated now. Clearing it first is done in the dashboard, which " +
	"installs the same template."

func planInstallAppTool() Tool {
	return planToolWith("plan_install_app", "Plan an install",
		"Plans POST /projects/{project}/{env}/apps/from-template: installing a template of the app store "+
			"into an env. The plan is the request and what preflight_install answers for it - every app it "+
			"would create, the app, its components and the apps it depends on, with names and images, and "+
			"what would stop it; a plan with issues cannot be applied. Applied, the endpoint answers the "+
			"ids of the apps created and of each one's first deployment, which is already queued. Nothing "+
			"is created until apply_plan.",
		NeedWrite, &applier{follow: "get_app_deployment with each deployment id shows it pull and start; " +
			"get_app_status shows the containers."},
		installInput,
		func(ctx context.Context, call *Call, in installArgs) (installPlan, *storedPlan, error) {
			ref, body, err := in.resolve(ctx, call)
			if err != nil {
				return installPlan{}, nil, err
			}
			result, err := preflight(ctx, call, ref, body)
			if err != nil {
				return installPlan{}, nil, err
			}
			out := installPlan{Project: ref.ProjectKey, Env: ref.Env, Request: body, Preflight: result.Data}
			if len(result.Data.Storage) > 0 {
				out.Warnings = append(out.Warnings, leftoverDataWarning)
			}
			if len(result.Data.Issues) > 0 {
				return out, nil, nil
			}
			raw, err := json.Marshal(body)
			if err != nil {
				return installPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/apps/from-template"), Body: raw,
				Summary: fmt.Sprintf("install %s as %s in %s/%s, %d app(s)", body.Template, body.Name,
					ref.ProjectKey, ref.Env, len(result.Data.Apps))}, nil
		})
}
