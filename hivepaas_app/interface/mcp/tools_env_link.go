package mcp

import (
	"context"
	"net/http"
	"net/url"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// The env vars that link an app to another - a database, a cache, a store -
// as the dashboard's Link to another app suggests them: references such as
// ${db.HIVEPAAS_HOST}, resolved when the app is deployed, rather than values
// copied in and stale on the next password change.

const (
	pathEnvLinkTargets     = "/env-vars/link-targets"
	pathEnvLinkSuggestions = "/env-vars/link-suggestions"
	pathEnvSelfSuggestions = "/env-vars/self-suggestions"
)

// EnvLinkEndpoints are the routes the env link tools use, as method and path
// under an app, so that the server's tests can find each in its router.
func EnvLinkEndpoints() []string {
	return []string{http.MethodGet + " " + pathEnvLinkTargets, http.MethodGet + " " + pathEnvLinkSuggestions,
		http.MethodGet + " " + pathEnvSelfSuggestions}
}

var listEnvLinkTargetsEndpoint = getEndpoint{
	name: "list_env_link_targets", title: "List what an app can link to",
	description: "GET /projects/{project}/{env}/apps/{app}" + pathEnvLinkTargets + ". The apps of the env " +
		"this app can link its env vars to - databases, caches, stores - each with its key, category and " +
		"engine. get_env_link_suggestions answers the env vars that link to one.",
	paths:  map[under]string{underApp: pathEnvLinkTargets},
	answer: func() any { return &appsettingsdto.ListEnvLinkTargetsResp{} },
}

var getEnvSelfSuggestionsEndpoint = getEndpoint{
	name: "get_env_self_suggestions", title: "Suggest an engine's own env vars",
	description: "GET /projects/{project}/{env}/apps/{app}" + pathEnvSelfSuggestions + ". The env vars an " +
		"engine's official image reads to set itself up - POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB for " +
		"PostgreSQL - each a reference such as ${HIVEPAAS_PASSWORD} to the credentials the app's own kind " +
		"settings publish, so they are kept in one place. For an app run from an engine's image by hand; an " +
		"app from the app store has them. For Redis and Valkey, which read none, a command to run instead. " +
		"warnings say what the kind settings lack; an engine with initOnly reads them only when it creates " +
		"its data. Give the vars to plan_update_app_settings, kind env-vars, and a command as deployment's command.",
	paths: map[under]string{underApp: pathEnvSelfSuggestions},
	query: &appsettingsdto.GetEnvSelfSuggestionsReq{},
	params: map[string]string{
		"engine": "the engine to suggest for, one of data.engines' ids; the app's App Kind engine when not given",
	},
	answer: func() any { return &appsettingsdto.GetEnvSelfSuggestionsResp{} },
}

// ---- get_env_link_suggestions ----

type envLinkInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id"`
	Env     string `json:"env" jsonschema:"the env's name, such as prod"`
	App     string `json:"app" jsonschema:"the key, name or id of the app that gets the env vars"`
	Target  string `json:"target" jsonschema:"the key, name or id of the app to link to; list_env_link_targets lists them"`
}

func getEnvLinkSuggestionsTool() Tool {
	return readTool("get_env_link_suggestions", "Suggest env vars linking to an app",
		"GET /projects/{project}/{env}/apps/{app}"+pathEnvLinkSuggestions+". The env vars that connect the "+
			"app to another of its env - a database, a cache, a store - in groups, as the dashboard's Link to "+
			"another app suggests them: a connection URL such as DATABASE_URL, the parts one by one, the "+
			"variables the engine's clients read (PGHOST, PGSSLMODE...), the recommended group marked. Each "+
			"value is a reference such as ${db.HIVEPAAS_PASSWORD}, resolved on each deployment, so a changed "+
			"password reaches the app; give them to plan_update_app_settings, kind env-vars, as they are. "+
			"warnings say what to fix first, such as a target with no port.",
		func(ctx context.Context, call *Call, in envLinkInput) (*appsettingsdto.GetEnvLinkSuggestionsResp, error) {
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return nil, err
			}
			apps, err := listApps(ctx, call, ref)
			if err != nil {
				return nil, err
			}
			byName := func(a listedApp) named { return named{id: a.ID, key: a.Key, name: a.Name} }
			app, err := pick(argApp, in.App, "list_apps", apps, byName)
			if err != nil {
				return nil, err
			}
			target, err := pick("target", in.Target, "list_env_link_targets", apps, byName)
			if err != nil {
				return nil, err
			}
			appRef := &appRef{envRef: *ref, AppID: app.ID, AppKey: app.Key, AppName: app.Name}

			var resp appsettingsdto.GetEnvLinkSuggestionsResp
			query := url.Values{"targetAppId": {target.ID}}
			if err = call.Get(ctx, appRef.path(pathEnvLinkSuggestions), query, &resp); err != nil {
				return nil, err
			}
			return &resp, nil
		})
}
