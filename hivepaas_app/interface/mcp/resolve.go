package mcp

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// A model names things the way a person does - "the api app of shop, in
// prod" - and the API takes ids. Resolution goes through the list endpoints as
// the caller, so what the caller may not see is not found, exactly as a name
// that does not exist. The lists are matched here rather than by the API's
// search, which reads names and notes but not keys.

// envRef is one env of one project, as the API's paths name it.
type envRef struct {
	ProjectID  string
	ProjectKey string
	Env        string
}

// path is the env's base path, followed by suffix.
func (r *envRef) path(suffix string) string {
	return "/projects/" + url.PathEscape(r.ProjectID) + "/" + url.PathEscape(r.Env) + suffix
}

// appRef is one app, as the API's paths name it.
type appRef struct {
	envRef
	AppID   string
	AppKey  string
	AppName string
}

// path is the app's base path, followed by suffix.
func (r *appRef) path(suffix string) string {
	return r.envRef.path("/apps/" + url.PathEscape(r.AppID) + suffix)
}

// listedApp is an app of an env's list, from the endpoint's own type, with the
// app it was made underneath, if any.
type listedApp struct {
	*appdto.AppResp
	// owner is the key of the app this one was made underneath, if any.
	owner string
}

// named is what a name is matched against.
type named struct {
	id, key, name string
}

func (n named) String() string {
	if n.name == "" || n.name == n.key {
		return n.key
	}
	return n.key + " (" + n.name + ")"
}

// pick finds the one item the input names: by id or key exactly, or else by
// name ignoring case. None and several are both errors a model can correct -
// several list the candidates, none says which tool lists them all.
func pick[T any](what, input, lister string, items []T, name func(T) named) (T, error) {
	var zero T
	input = strings.TrimSpace(input)
	if input == "" {
		return zero, &InputError{Message: what + " is required; " + lister + " lists them"}
	}
	var exact, byName []T
	for _, item := range items {
		n := name(item)
		switch {
		case n.id == input || n.key == input:
			exact = append(exact, item)
		case strings.EqualFold(n.name, input):
			byName = append(byName, item)
		}
	}
	if len(exact) == 0 {
		exact = byName
	}
	switch len(exact) {
	case 1:
		return exact[0], nil
	case 0:
		return zero, &InputError{Message: fmt.Sprintf(
			"no %s %q is visible to the API key's user; %s lists them", what, input, lister)}
	}
	candidates := make([]string, 0, len(exact))
	for _, item := range exact {
		candidates = append(candidates, name(item).String())
	}
	return zero, &InputError{Message: fmt.Sprintf("%q names %d of the %ss: %s; name one by its key",
		input, len(exact), what, strings.Join(candidates, ", "))}
}

// resolveEnv finds a project, then one of its envs.
func resolveEnv(ctx context.Context, call *Call, project, env string) (*envRef, error) {
	projects, err := listProjects(ctx, call, "")
	if err != nil {
		return nil, err
	}
	p, err := pick(argProject, project, "list_projects", slices.DeleteFunc(projects, isNil),
		func(p *projectdto.ProjectResp) named {
			return named{id: p.ID, key: p.Key, name: p.Name}
		})
	if err != nil {
		return nil, err
	}
	e, err := pick("env of "+p.Key, env, "list_projects", slices.DeleteFunc(p.Envs, isNil),
		func(e *projectdto.ProjectEnvResp) named {
			return named{id: e.ID, key: e.Name, name: e.Name}
		})
	if err != nil {
		return nil, err
	}
	return &envRef{ProjectID: p.ID, ProjectKey: p.Key, Env: e.Name}, nil
}

// resolveApp finds a project, one of its envs, then one of the env's apps.
func resolveApp(ctx context.Context, call *Call, project, env, app string) (*appRef, error) {
	ref, err := resolveEnv(ctx, call, project, env)
	if err != nil {
		return nil, err
	}
	apps, err := listApps(ctx, call, ref, false)
	if err != nil {
		return nil, err
	}
	a, err := pick(argApp, app, "list_apps", apps, func(a listedApp) named {
		return named{id: a.ID, key: a.Key, name: a.Name}
	})
	if err != nil {
		return nil, err
	}
	return &appRef{envRef: *ref, AppID: a.ID, AppKey: a.Key, AppName: a.Name}, nil
}

// listApps asks the env's app list, as the caller: every app, those that
// belong to another app included, the way a name may refer to any of them.
func listApps(ctx context.Context, call *Call, ref *envRef, stats bool) ([]listedApp, error) {
	query := url.Values{paramPageLimit: {strconv.Itoa(maxListed)}, "getChildApps": {paramTrue}}
	if stats {
		query.Set(paramGetStats, paramTrue)
	}
	var resp appdto.ListAppResp
	if err := call.Get(ctx, ref.path("/apps"), query, &resp); err != nil {
		return nil, err
	}
	// An app made underneath another - a preview, a template's component - is
	// listed inside its owner. The list is made flat, each child naming its owner.
	var flat []listedApp
	for _, a := range resp.Data {
		if a == nil {
			continue
		}
		flat = append(flat, listedApp{AppResp: a})
		for _, child := range slices.Concat(a.ChildApps, a.LogicalChildApps) {
			if child != nil {
				flat = append(flat, listedApp{AppResp: child, owner: a.Key})
			}
		}
	}
	return flat, nil
}

// isNil is for dropping the nil entries a list of pointers may have.
func isNil[T any](v *T) bool { return v == nil }
