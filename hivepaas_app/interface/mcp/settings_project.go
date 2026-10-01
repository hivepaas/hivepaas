package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

// A project's settings, and an env's, are kinds too, each an endpoint pair -
// or a GET alone - under the project or the env: read and changed as an app's
// are, a kind at a time, through the kind's own endpoints and types.

// projectSettingsKind is one kind of the settings a project or an env keeps.
type projectSettingsKind struct {
	name string
	// path is the endpoint pair's, under the project or the env.
	path string
	// newGet is the GET's data type where the kind is kept; newUpdate the
	// PUT's request, where the tools may change it.
	newGet    map[under]func() any
	newUpdate map[under]func() any
}

var projectSettingsKinds = []projectSettingsKind{
	{name: "env-vars", path: "/env-vars",
		newGet: map[under]func() any{
			underProject: func() any { return &projectsettingsdto.EnvVarsResp{} },
			underEnv:     func() any { return &projectenvsettingsdto.EnvVarsResp{} },
		},
		newUpdate: map[under]func() any{
			underProject: func() any { return &projectsettingsdto.UpdateProjectEnvVarsReq{} },
			underEnv:     func() any { return &projectenvsettingsdto.UpdateProjectEnvEnvVarsReq{} },
		}},
	{name: paramDomain, path: "/domain-settings",
		newGet: map[under]func() any{
			underProject: func() any { return &domainsettingsdto.DomainSettingsResp{} },
		}},
}

func projectSettingsKindNames() string {
	names := make([]string, 0, len(projectSettingsKinds))
	for _, kind := range projectSettingsKinds {
		names = append(names, kind.name)
	}
	return strings.Join(names, ", ")
}

func findProjectSettingsKind(name string) (*projectSettingsKind, error) {
	name = strings.TrimSpace(name)
	for i := range projectSettingsKinds {
		if projectSettingsKinds[i].name == name {
			return &projectSettingsKinds[i], nil
		}
	}
	return nil, &InputError{Message: fmt.Sprintf("no settings kind %q; the kinds are %s", name,
		projectSettingsKindNames())}
}

// ProjectSettingsEndpoints are the routes the project settings tools use, as
// method and path under a project or under an env, so that the server's tests
// can find each in its router.
func ProjectSettingsEndpoints() (underProjectRoutes, underEnvRoutes []string) {
	for _, kind := range projectSettingsKinds {
		for u, routes := range map[under]*[]string{underProject: &underProjectRoutes, underEnv: &underEnvRoutes} {
			if kind.newGet[u] != nil {
				*routes = append(*routes, http.MethodGet+" "+kind.path)
			}
			if kind.newUpdate[u] != nil {
				*routes = append(*routes, http.MethodPut+" "+kind.path)
			}
		}
	}
	return underProjectRoutes, underEnvRoutes
}

// projectSettingsPlace is where a project settings tool goes: an env when one
// is named, else the project.
type projectSettingsPlace struct {
	under      under
	path       string
	projectKey string
	env        string
}

func (p *projectSettingsPlace) String() string {
	if p.env != "" {
		return p.projectKey + "/" + p.env
	}
	return p.projectKey
}

func resolveProjectSettingsPlace(ctx context.Context, call *Call, project, env string) (*projectSettingsPlace,
	error) {
	if strings.TrimSpace(env) != "" {
		ref, err := resolveEnv(ctx, call, project, env)
		if err != nil {
			return nil, err
		}
		return &projectSettingsPlace{under: underEnv, path: ref.path(""), projectKey: ref.ProjectKey,
			env: ref.Env}, nil
	}
	ref, err := resolveProject(ctx, call, project)
	if err != nil {
		return nil, err
	}
	return &projectSettingsPlace{under: underProject, path: ref.path(""), projectKey: ref.ProjectKey}, nil
}

// kept says where a kind is kept, for a model that asked for it elsewhere.
func (k *projectSettingsKind) kept() string {
	if k.newGet[underEnv] == nil {
		return "a project's: leave env out"
	}
	return "a project's and each env's"
}

// ---- get_project_settings ----

type projectSettingsInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env,omitempty" jsonschema:"the env's name, for its own settings; left out, the project's"`
	Kind    string `json:"kind" jsonschema:"one of the kinds the tool's description names, such as env-vars"`
}

type projectSettingsAnswer struct {
	Project string `json:"project"`
	Env     string `json:"env,omitempty"`
	Kind    string `json:"kind"`
	// Settings are the endpoint's data, as it answers them.
	Settings any `json:"settings"`
}

func getProjectSettingsTool() Tool {
	return readTool("get_project_settings", "Read a project's or an env's settings",
		"Reads one kind of the settings a project keeps for all its apps, or an env for its own: "+
			"env-vars (the env vars every app of the project, or of the env, gets - an env's answer "+
			"shows those it inherits from its project too) or domain (a project's root domain, the "+
			"domains its apps may use, and how their certificates are made). Given an env, the env's; "+
			"left out, the project's. An app's own settings are get_app_settings'.",
		func(ctx context.Context, call *Call, in projectSettingsInput) (projectSettingsAnswer, error) {
			kind, err := findProjectSettingsKind(in.Kind)
			if err != nil {
				return projectSettingsAnswer{}, err
			}
			at, err := resolveProjectSettingsPlace(ctx, call, in.Project, in.Env)
			if err != nil {
				return projectSettingsAnswer{}, err
			}
			settings, err := readProjectSettings(ctx, call, at, kind)
			if err != nil {
				return projectSettingsAnswer{}, err
			}
			return projectSettingsAnswer{Project: at.projectKey, Env: at.env, Kind: kind.name,
				Settings: settings}, nil
		})
}

func readProjectSettings(ctx context.Context, call *Call, at *projectSettingsPlace, kind *projectSettingsKind) (
	map[string]any, error) {
	newGet := kind.newGet[at.under]
	if newGet == nil {
		return nil, &InputError{Message: fmt.Sprintf("%s settings are %s", kind.name, kind.kept())}
	}
	var resp struct {
		Data json.RawMessage `json:"data"`
	}
	if err := call.Get(ctx, at.path+kind.path, nil, &resp); err != nil {
		return nil, err
	}
	settings, err := through(resp.Data, newGet())
	if err != nil {
		return nil, fmt.Errorf("mcp: reading %s settings: %w", kind.name, err)
	}
	return settings, nil
}

// ---- plan_update_project_settings ----

type updateProjectSettingsInput struct {
	Project string         `json:"project" jsonschema:"the project's key, name or id"`
	Env     string         `json:"env,omitempty" jsonschema:"the env's name; left out, the project's"`
	Kind    string         `json:"kind" jsonschema:"the kind of settings, as get_project_settings takes it"`
	Changes map[string]any `json:"changes" jsonschema:"a JSON merge patch of get_project_settings' answer"`
}

type projectSettingsPlan struct {
	Project string        `json:"project"`
	Env     string        `json:"env,omitempty"`
	Kind    string        `json:"kind"`
	Changes []fieldChange `json:"changes"`
	// Ignored are fields of the patch the endpoint does not take.
	Ignored []string `json:"ignored,omitempty"`
}

func planUpdateProjectSettingsTool() Tool {
	return planTool("plan_update_project_settings", "Plan a project or env settings change",
		"Plans changing the env vars a project gives all its apps, or an env its own: give the fields "+
			"to change as a merge patch of what get_project_settings answered; a list is replaced whole, "+
			"and null removes a field. Answers each value that would change, before and after. The apps "+
			"get the change on their next deployment. A change the endpoint refuses is refused when the "+
			"plan is applied, in its words. Nothing changes until apply_plan.",
		NeedWrite, &applier{refused: settingsRefused,
			follow: "get_project_settings reads them as they are now; plan_redeploy_app brings an app the change."},
		func(ctx context.Context, call *Call, in updateProjectSettingsInput) (projectSettingsPlan, *storedPlan,
			error) {
			kind, err := findProjectSettingsKind(in.Kind)
			if err != nil {
				return projectSettingsPlan{}, nil, err
			}
			delete(in.Changes, updateVerField)
			if len(in.Changes) == 0 {
				return projectSettingsPlan{}, nil, &InputError{Message: "changes is empty: give the fields to change"}
			}
			at, err := resolveProjectSettingsPlace(ctx, call, in.Project, in.Env)
			if err != nil {
				return projectSettingsPlan{}, nil, err
			}
			newUpdate := kind.newUpdate[at.under]
			if newUpdate == nil {
				return projectSettingsPlan{}, nil, &InputError{Message: fmt.Sprintf(
					"%s settings are not changed through these tools; the dashboard changes them", kind.name)}
			}
			current, err := readProjectSettings(ctx, call, at, kind)
			if err != nil {
				return projectSettingsPlan{}, nil, err
			}
			out := projectSettingsPlan{Project: at.projectKey, Env: at.env, Kind: kind.name}
			diff, err := diffSettings(kind.name, newUpdate, current, in.Changes)
			if err != nil {
				return projectSettingsPlan{}, nil, err
			}
			out.Changes, out.Ignored = diff.changes, diff.ignored
			if diff.body == nil {
				return out, nil, nil
			}
			return out, &storedPlan{Method: http.MethodPut, Path: at.path + kind.path, Body: diff.body,
					Summary: fmt.Sprintf("change %s settings of %s: %s", kind.name, at, describePaths(diff.paths()))},
				nil
		})
}
