package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// ---- plan_create_app ----

type createAppInput struct {
	Project string   `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string   `json:"env" jsonschema:"the env's name, such as production"`
	Name    string   `json:"name" jsonschema:"the new app's name, up to 100 characters; its key is made from it"`
	Note    string   `json:"note,omitempty" jsonschema:"a note on the app, shown with it"`
	Tags    []string `json:"tags,omitempty" jsonschema:"tags to file the app under"`
}

type createAppPlan struct {
	Project string `json:"project"`
	Env     string `json:"env"`
	// Request is the create endpoint's request the plan sends.
	Request *appdto.AppBaseReq `json:"request"`
	Next    string             `json:"next"`
}

const createAppNext = "The app is created empty, as the dashboard's New App does: nothing runs until it has " +
	"a source and is deployed. Then plan_update_app_settings - kind deployment for its image " +
	"(activeMethod image, imageSource.image), kind routing for its port and a domain, kind env-vars for its " +
	"variables - and plan_redeploy_app to start it."

func planCreateAppTool() Tool {
	return planTool("plan_create_app", "Plan creating an app",
		"Plans POST /projects/{project}/{env}/apps: a new, empty app in an env, as the dashboard's New App "+
			"creates one - to run an image or a repository of the person's own, rather than an app store "+
			"template (plan_install_app). A name another app of the env has is refused here. Applied, the "+
			"endpoint answers the new app's id; its source, port and variables are then set with "+
			"plan_update_app_settings, and plan_redeploy_app starts it. Nothing is created until apply_plan.",
		NeedWrite, &applier{follow: "get_app with the new id; then plan_update_app_settings and " +
			"plan_redeploy_app, as the plan's next says."},
		func(ctx context.Context, call *Call, in createAppInput) (createAppPlan, *storedPlan, error) {
			name := strings.TrimSpace(in.Name)
			if name == "" {
				return createAppPlan{}, nil, &InputError{Message: "name is required"}
			}
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return createAppPlan{}, nil, err
			}
			apps, err := listApps(ctx, call, ref, false)
			if err != nil {
				return createAppPlan{}, nil, err
			}
			for _, a := range apps {
				if strings.EqualFold(a.Name, name) || strings.EqualFold(a.Key, name) {
					return createAppPlan{}, nil, &InputError{Message: fmt.Sprintf(
						"%s/%s has an app named %s already (%s): choose another name, or change that one",
						ref.ProjectKey, ref.Env, name, a.Key)}
				}
			}

			body := &appdto.AppBaseReq{Name: name, Status: base.AppStatusActive, Note: in.Note, Tags: in.Tags}
			raw, err := json.Marshal(body)
			if err != nil {
				return createAppPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			out := createAppPlan{Project: ref.ProjectKey, Env: ref.Env, Request: body, Next: createAppNext}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/apps"), Body: raw,
				Summary: fmt.Sprintf("create the app %s in %s/%s", name, ref.ProjectKey, ref.Env)}, nil
		})
}
