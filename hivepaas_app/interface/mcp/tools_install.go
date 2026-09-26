package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type plannedApp struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Kind     string `json:"kind"`
	Role     string `json:"role,omitempty"`
	Template string `json:"template"`
	Version  string `json:"version,omitempty"`
	Image    string `json:"image"`
}

type installPlan struct {
	Project  string             `json:"project"`
	Env      string             `json:"env"`
	Template string             `json:"template"`
	Apps     []plannedApp       `json:"apps"`
	Issues   []preflightIssue   `json:"issues"`
	Storage  []preflightStorage `json:"leftoverData,omitempty"`
	Warnings []string           `json:"warnings,omitempty"`
}

const leftoverDataWarning = "Some of these apps would find data an earlier install left in their " +
	"directories, and start with it: a database keeps its old data and the password it was created with. " +
	"Clearing it first is done in the dashboard, which installs the same template."

func planInstallAppTool() Tool {
	return planTool("plan_install_app", "Plan an install",
		"Plans installing a template of the app store into an env: every app it would create - the app, "+
			"its components, the apps it depends on - with names and images, and what would stop it. A "+
			"secret parameter left out is generated. Takes what preflight_install takes. Nothing is "+
			"created until apply_plan.",
		NeedWrite, &applier{follow: "get_app and get_app_status on each app created show its first deployment.",
			result: installResult},
		func(ctx context.Context, call *Call, in preflightInput) (installPlan, *storedPlan, error) {
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return installPlan{}, nil, err
			}
			body := in.body()
			var resp struct {
				Data struct {
					Storage []apiPreflightStorage `json:"storage"`
					Issues  []preflightIssue      `json:"issues"`
					Apps    []plannedApp          `json:"apps"`
				} `json:"data"`
			}
			if err = call.Post(ctx, ref.path("/apps/from-template/preflight"), body, &resp); err != nil {
				return installPlan{}, nil, err
			}
			out := installPlan{Project: ref.ProjectKey, Env: ref.Env, Template: in.Template,
				Apps: resp.Data.Apps, Issues: resp.Data.Issues, Storage: makePreflightStorage(resp.Data.Storage)}
			if out.Issues == nil {
				out.Issues = []preflightIssue{}
			}
			if len(out.Storage) > 0 {
				out.Warnings = append(out.Warnings, leftoverDataWarning)
			}
			if len(out.Issues) > 0 {
				return out, nil, nil
			}
			raw, err := json.Marshal(body)
			if err != nil {
				return installPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/apps/from-template"), Body: raw,
				Summary: fmt.Sprintf("install %s as %s in %s/%s, %d app(s)", in.Template, in.Name,
					ref.ProjectKey, ref.Env, len(out.Apps))}, nil
		})
}

// installResult is the created apps by their ids, which is what the create
// answers: a model finds them by name with list_apps.
func installResult(data json.RawMessage) (any, error) {
	type idRef struct {
		ID string `json:"id"`
	}
	type role struct {
		Name string `json:"name"`
		App  *idRef `json:"app"`
	}
	var created struct {
		App          *idRef `json:"app"`
		Deployment   *idRef `json:"deployment"`
		Components   []role `json:"components"`
		Dependencies []role `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("mcp: decoding the answer: %w", err)
	}
	out := map[string]any{}
	if created.App != nil {
		out["app"] = created.App.ID
	}
	if created.Deployment != nil {
		out["deployment"] = created.Deployment.ID
	}
	for _, c := range created.Components {
		if c.App != nil {
			out["component:"+c.Name] = c.App.ID
		}
	}
	for _, d := range created.Dependencies {
		if d.App != nil {
			out["dependency:"+d.Name] = d.App.ID
		}
	}
	return out, nil
}
