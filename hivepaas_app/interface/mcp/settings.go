package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

// An app's settings are kinds, each an endpoint pair of appsettingsuc: a GET
// that answers them and a PUT that takes them. The tools read and change one
// kind at a time, through the kind's own endpoints and types. See
// docs/superpowers/specs/2026-09-26-mcp-tools-follow-the-api-design.md.

// settingsKind is one kind of an app's settings, as the API serves it.
type settingsKind struct {
	name string
	// path is the endpoint pair's, under the app.
	path string
	// newGet is the GET's data type, newUpdate the PUT's request: a body goes
	// through the request type, so what the PUT does not take is dropped by
	// the API's own type rather than a list kept here.
	newGet    func() any
	newUpdate func() any
}

var appSettingsKinds = []settingsKind{
	{"env-vars", "/env-vars",
		func() any { return &appsettingsdto.EnvVarsResp{} },
		func() any { return &appsettingsdto.UpdateAppEnvVarsReq{} }},
	{"deployment", "/deployment-settings",
		func() any { return &appsettingsdto.DeploymentSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppDeploymentSettingsReq{} }},
	{"routing", "/routing-settings",
		func() any { return &appsettingsdto.RoutingSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppRoutingSettingsReq{} }},
	{"service", "/service-settings",
		func() any { return &appsettingsdto.ServiceSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppServiceSettingsReq{} }},
	{"network", "/network-settings",
		func() any { return &appsettingsdto.NetworkSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppNetworkSettingsReq{} }},
	{"resource", "/resource-settings",
		func() any { return &appsettingsdto.ResourceSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppResourceSettingsReq{} }},
	{"container", "/container-settings",
		func() any { return &appsettingsdto.ContainerSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppContainerSettingsReq{} }},
	{"storage", "/storage-settings",
		func() any { return &appsettingsdto.StorageSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppStorageSettingsReq{} }},
	{"feature", "/feature-settings",
		func() any { return &appfeaturesettingsdto.AppFeatureSettingsResp{} },
		func() any { return &appfeaturesettingsdto.UpdateAppFeatureSettingsReq{} }},
	{"kind", "/kind-settings",
		func() any { return &appsettingsdto.AppKindSettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppKindSettingsReq{} }},
	{"docker-api", "/docker-api-settings",
		func() any { return &appsettingsdto.AppDockerAPISettingsResp{} },
		func() any { return &appsettingsdto.UpdateAppDockerAPISettingsReq{} }},
}

func settingsKindNames() string {
	names := make([]string, 0, len(appSettingsKinds))
	for _, kind := range appSettingsKinds {
		names = append(names, kind.name)
	}
	return strings.Join(names, ", ")
}

func findSettingsKind(name string) (*settingsKind, error) {
	name = strings.TrimSpace(name)
	for i := range appSettingsKinds {
		if appSettingsKinds[i].name == name {
			return &appSettingsKinds[i], nil
		}
	}
	return nil, &InputError{Message: fmt.Sprintf("no settings kind %q; the kinds are %s", name, settingsKindNames())}
}

// SettingsEndpoints are the routes the settings tools use, as method and path
// under an app, so that the server's tests can find each in its router.
func SettingsEndpoints() []string {
	var out []string
	for _, kind := range appSettingsKinds {
		out = append(out, http.MethodGet+" "+kind.path, http.MethodPut+" "+kind.path)
	}
	return out
}

// ---- get_app_settings ----

type settingsInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the app's key, name or id"`
	Kind    string `json:"kind" jsonschema:"one of the kinds the tool's description names, such as env-vars"`
}

type settingsAnswer struct {
	App  string `json:"app"`
	Kind string `json:"kind"`
	// Settings are the endpoint's data, as it answers them.
	Settings any `json:"settings"`
}

func getAppSettingsTool() Tool {
	return readTool("get_app_settings", "Read an app's settings",
		"Reads one kind of an app's settings, as the dashboard's page for it does: env-vars, deployment "+
			"(its source: image, or repository and branch; and its command), routing (the container port, "+
			"whether the app is exposed, and its domains - HTTP ones, and the TCP ones that reach a database "+
			"or cache from outside, each with its certificate, TLS passthrough and extraAlpnProtocols), "+
			"service (replicas), network (published ports), resource (CPU and memory), container, storage "+
			"(mounts), feature, kind (the database or cache it is: engine, version, credentials, sslMode - "+
			"not its domains, which are routing's) or docker-api. Secrets are masked. Only the kind asked for "+
			"is read. To connect an app to a database, get_env_link_suggestions answers the env vars; for an "+
			"engine's image to set itself up, get_env_self_suggestions.",
		func(ctx context.Context, call *Call, in settingsInput) (settingsAnswer, error) {
			kind, err := findSettingsKind(in.Kind)
			if err != nil {
				return settingsAnswer{}, err
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return settingsAnswer{}, err
			}
			settings, err := readSettings(ctx, call, ref, kind)
			if err != nil {
				return settingsAnswer{}, err
			}
			return settingsAnswer{App: ref.AppKey, Kind: kind.name, Settings: settings}, nil
		})
}

// readSettings is a kind's GET: its data through the GET's own type, as JSON.
func readSettings(ctx context.Context, call *Call, ref *appRef, kind *settingsKind) (map[string]any, error) {
	var resp struct {
		Data json.RawMessage `json:"data"`
	}
	if err := call.Get(ctx, ref.path(kind.path), nil, &resp); err != nil {
		return nil, err
	}
	settings, err := through(resp.Data, kind.newGet())
	if err != nil {
		return nil, fmt.Errorf("mcp: reading %s settings: %w", kind.name, err)
	}
	return settings, nil
}

// through decodes JSON into one of the API's types and encodes it back: what
// the type does not have is dropped, and what it has is in its own form.
func through(raw []byte, typed any) (map[string]any, error) {
	if err := json.Unmarshal(raw, typed); err != nil {
		return nil, err //nolint:wrapcheck // the caller words it
	}
	encoded, err := json.Marshal(typed)
	if err != nil {
		return nil, err //nolint:wrapcheck // the caller words it
	}
	out := map[string]any{}
	if err = json.Unmarshal(encoded, &out); err != nil {
		return nil, err //nolint:wrapcheck // the caller words it
	}
	return out, nil
}

// ---- plan_update_app_settings ----

type updateSettingsInput struct {
	Project string         `json:"project" jsonschema:"the project's key, name or id"`
	Env     string         `json:"env" jsonschema:"the env's name, such as prod"`
	App     string         `json:"app" jsonschema:"the app's key, name or id"`
	Kind    string         `json:"kind" jsonschema:"the kind of settings, as get_app_settings takes it"`
	Changes map[string]any `json:"changes" jsonschema:"a JSON merge patch of get_app_settings' answer"`
}

type settingsPlan struct {
	App     string        `json:"app"`
	Project string        `json:"project"`
	Env     string        `json:"env"`
	Kind    string        `json:"kind"`
	Changes []fieldChange `json:"changes"`
	// Ignored are fields of the patch the endpoint does not take.
	Ignored []string `json:"ignored,omitempty"`
}

const updateVerField = "updateVer"

func planUpdateAppSettingsTool() Tool {
	return planTool("plan_update_app_settings", "Plan a settings change",
		"Plans changing one kind of an app's settings: give the fields to change as a merge patch of what "+
			"get_app_settings answered; a list is replaced whole, and null removes a field. Answers each "+
			"value that would change, before and after. A masked "+
			"secret sent back as it is keeps the secret. A change the endpoint refuses is refused when the "+
			"plan is applied, in its words. Nothing changes until apply_plan.",
		NeedWrite, &applier{refused: settingsRefused,
			follow: "get_app_settings reads them as they are now; get_app_status shows a restart they cause."},
		func(ctx context.Context, call *Call, in updateSettingsInput) (settingsPlan, *storedPlan, error) {
			kind, err := findSettingsKind(in.Kind)
			if err != nil {
				return settingsPlan{}, nil, err
			}
			delete(in.Changes, updateVerField)
			if len(in.Changes) == 0 {
				return settingsPlan{}, nil, &InputError{Message: "changes is empty: give the fields to change"}
			}
			ref, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return settingsPlan{}, nil, err
			}
			current, err := readSettings(ctx, call, ref, kind)
			if err != nil {
				return settingsPlan{}, nil, err
			}
			return makeSettingsPlan(ref, kind, current, in.Changes)
		})
}

func makeSettingsPlan(ref *appRef, kind *settingsKind, current, changes map[string]any) (
	settingsPlan, *storedPlan, error) {
	out := settingsPlan{App: ref.AppKey, Project: ref.ProjectKey, Env: ref.Env, Kind: kind.name}
	diff, err := diffSettings(kind.name, kind.newUpdate, current, changes)
	if err != nil {
		return settingsPlan{}, nil, err
	}
	out.Changes, out.Ignored = diff.changes, diff.ignored
	if diff.body == nil {
		return out, nil, nil
	}
	return out, &storedPlan{Method: http.MethodPut, Path: ref.path(kind.path), Body: diff.body,
		Summary: fmt.Sprintf("change %s settings of %s in %s/%s: %s", kind.name, ref.AppKey, ref.ProjectKey,
			ref.Env, describePaths(diff.paths()))}, nil
}

// settingsDiff is what a settings change would do: each value that changes,
// the fields of the patch the PUT does not take, and the PUT's body - nil when
// nothing changes.
type settingsDiff struct {
	changes []fieldChange
	ignored []string
	body    []byte
}

func (d settingsDiff) paths() []string {
	paths := make([]string, 0, len(d.changes))
	for _, change := range d.changes {
		paths = append(paths, change.Path)
	}
	return paths
}

// diffSettings patches settings read as current, and puts both through the
// PUT's request type: what the PUT does not take falls away, and what is left
// is compared.
func diffSettings(kindName string, newUpdate func() any, current, changes map[string]any) (settingsDiff, error) {
	currentRaw, err := json.Marshal(current)
	if err != nil {
		return settingsDiff{}, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	before, err := through(currentRaw, newUpdate())
	if err != nil {
		return settingsDiff{}, fmt.Errorf("mcp: %s settings do not fit their own update: %w", kindName, err)
	}
	patchedRaw, err := json.Marshal(mergePatch(current, changes))
	if err != nil {
		return settingsDiff{}, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	after, err := through(patchedRaw, newUpdate())
	if err != nil {
		return settingsDiff{}, &InputError{Message: fmt.Sprintf(
			"the changes do not fit the %s settings: %v", kindName, err)}
	}
	diff := settingsDiff{changes: diffFields(before, after), ignored: droppedFields(changes, after)}
	if len(diff.changes) == 0 {
		diff.changes = []fieldChange{}
		return diff, nil
	}
	if diff.body, err = json.Marshal(after); err != nil {
		return settingsDiff{}, fmt.Errorf("mcp: encoding a plan: %w", err)
	}
	return diff, nil
}

// settingsRefused words the refusal of settings changed since the plan: every
// settings update checks the updateVer the plan read.
func settingsRefused(apiErr *APIError) error {
	if apiErr.Code == codeUpdateVerMismatched {
		return planMoved("the settings")
	}
	return nil
}

const codeUpdateVerMismatched = "ERR_UPDATE_VER_MISMATCHED"
