package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

// ---- plan_create_preview ----

type createPreviewInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string `json:"env" jsonschema:"the env's name, such as production"`
	App     string `json:"app" jsonschema:"the app to preview, by key, name or id; it deploys from a repository"`
	//nolint:lll // a schema description is one line
	RepoRef string `json:"repoRef" jsonschema:"what to deploy: a pull request - pull/42 on GitHub and Gitea, merge-requests/42 on GitLab - or a branch"`
	//nolint:lll // a schema description is one line
	Subdomain string `json:"subdomain,omitempty" jsonschema:"the prefix of the preview's domains; pr-<number> for a pull request when not given"`
	NoStart   bool   `json:"noStart,omitempty" jsonschema:"true to create the preview without starting it"`
	//nolint:lll // a schema description is one line
	CloneDBApps *bool `json:"cloneDbApps,omitempty" jsonschema:"true for copies of the database apps its preview settings name, false for none; their setting when not given"`
}

type createPreviewPlan struct {
	Project   string `json:"project"`
	Env       string `json:"env"`
	App       string `json:"app"`
	RepoURL   string `json:"repoUrl"`
	RepoRef   string `json:"repoRef"`
	Subdomain string `json:"subdomain"`
	Start     bool   `json:"start"`
	Databases string `json:"databases"`
	// WithheldSecrets are the app's secrets the preview goes without, not
	// being inheritable, and the variables empty in it for that.
	WithheldSecrets []*apppreviewdto.WithheldSecretResp `json:"withheldSecrets,omitempty"`
	Effect          string                              `json:"effect"`
}

func planCreatePreviewTool() Tool {
	return planTool("plan_create_preview", "Plan a preview of an app",
		"Plans POST /projects/{project}/{env}/apps/{app}/previews: a preview - a copy of the app, made under "+
			"it in the same env, built from a pull request or a branch - as the dashboard's Create a preview "+
			"does. The app must deploy from a repository and have App Preview on in its feature settings. The "+
			"preview's domains are the app's, prefixed: pr-42-shop.example.com for pull request 42 of "+
			"shop.example.com. The plan asks the app's preview settings first, and says which of the app's "+
			"secrets the preview goes without - those not inheritable - and the variables empty in it for that. "+
			"Applied, the endpoint answers the task that makes the preview. Nothing is made until apply_plan.",
		NeedWrite, &applier{follow: "get_task with the task's id follows the preview being made; then " +
			"list_app_previews lists it, and the app tools read it by its key."},
		func(ctx context.Context, call *Call, in createPreviewInput) (createPreviewPlan, *storedPlan, error) {
			repoRef := strings.TrimSpace(in.RepoRef)
			if repoRef == "" {
				return createPreviewPlan{}, nil, &InputError{Message: "repoRef is required: a pull request, such " +
					"as pull/42, or a branch"}
			}
			app, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return createPreviewPlan{}, nil, err
			}
			var prep apppreviewdto.PrepareCreatePreviewResp
			if err = call.Post(ctx, app.path("/previews/prepare"), struct{}{}, &prep); err != nil {
				return createPreviewPlan{}, nil, err
			}
			p := prep.Data
			if p == nil || !p.Enabled {
				return createPreviewPlan{}, nil, &InputError{Message: "previews are off for " + app.AppKey + ": " +
					"turn on App Preview in its feature settings (plan_update_app_settings, kind feature, " +
					"previewSettings.enabled)"}
			}

			databases := "none: the app's preview settings name no database app to clone"
			cloneDBApps := in.CloneDBApps
			switch {
			case p.CanSkipCloningDBApps: // cloned unless told not to
				databases = "copies of the database apps its preview settings name"
				if cloneDBApps != nil && !*cloneDBApps {
					databases = "none: told not to clone the database apps; it shares the app's"
				}
			case p.CanCloneDBApps: // not cloned unless told to
				databases = "none: it shares the app's, as its preview settings say"
				if cloneDBApps != nil && *cloneDBApps {
					databases = "copies of the database apps its preview settings name"
				}
			case cloneDBApps != nil && *cloneDBApps:
				return createPreviewPlan{}, nil, &InputError{Message: app.AppKey + "'s preview settings name no " +
					"database app to clone: leave cloneDbApps out"}
			default:
				cloneDBApps = nil
			}

			body := &apppreviewdto.CreatePreviewReq{RepoRef: repoRef,
				CustomSubdomain: strings.ToLower(strings.TrimSpace(in.Subdomain)), NoStart: in.NoStart,
				CloneDBApps: cloneDBApps}
			raw, err := json.Marshal(body)
			if err != nil {
				return createPreviewPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			subdomain := body.CustomSubdomain
			if subdomain == "" {
				subdomain = "pr-<number> for a pull request"
			}
			out := createPreviewPlan{
				Project: app.ProjectKey, Env: app.Env, App: app.AppKey, RepoURL: p.RepoURL, RepoRef: repoRef,
				Subdomain: subdomain, Start: !in.NoStart, Databases: databases, WithheldSecrets: p.WithheldSecrets,
				Effect: "a preview app is made under " + app.AppKey + " from " + repoRef + ", with its settings, " +
					"variables and inheritable secrets, and deployed - or only made, with noStart",
			}
			return out, &storedPlan{Method: http.MethodPost, Path: app.path("/previews"), Body: raw,
				Summary: fmt.Sprintf("preview %s of %s in %s/%s", repoRef, app.AppKey, app.ProjectKey, app.Env)}, nil
		})
}
