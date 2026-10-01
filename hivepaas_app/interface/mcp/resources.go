package mcp

import (
	"context"
	_ "embed"
	"errors"
	"net/http"
	"net/url"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Resources are documents a client may attach to a conversation; prompts are
// starting points it offers a person. Neither grants anything a tool does not:
// a template is read as the caller, and a prompt only arranges tool calls.

const (
	templateURIPrefix = "hivepaas://templates/"
	dockerAPIURI      = "hivepaas://docs/docker-api"
	databasesURI      = "hivepaas://docs/databases"
	markdownMIME      = "text/markdown"
)

// dockerAPIGuide is the dashboard's permissions guide, as text. Kept beside
// pkg/dockerproxy and changed with it, as the dashboard's copy is.
//
//go:embed docs/docker-api.md
var dockerAPIGuide string

// databasesGuide is how HivePaaS runs, connects and exposes databases, as the
// tools reach it: the prompts about databases send a model here.
//
//go:embed docs/databases.md
var databasesGuide string

func addResources(s *mcpsdk.Server, deps *Deps) {
	s.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: templateURIPrefix + "{name}",
		Name:        argTemplate,
		Title:       "App store template",
		Description: "A template's description, as the app store shows it.",
		MIMEType:    markdownMIME,
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return readTemplateResource(ctx, deps, req.Params.URI)
	})

	s.AddResource(&mcpsdk.Resource{
		URI:         dockerAPIURI,
		Name:        "docker-api",
		Title:       "What an app may do through the Docker API",
		Description: "The rules the Docker API proxy holds an app to, and what each permission opens.",
		MIMEType:    markdownMIME,
	}, func(_ context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return &mcpsdk.ReadResourceResult{Contents: []*mcpsdk.ResourceContents{
			{URI: req.Params.URI, MIMEType: markdownMIME, Text: dockerAPIGuide},
		}}, nil
	})

	s.AddResource(&mcpsdk.Resource{
		URI:         databasesURI,
		Name:        "databases",
		Title:       "Databases on HivePaaS",
		Description: "Connecting an app to a database, running one from its image, reaching it from outside, SSL mode.",
		MIMEType:    markdownMIME,
	}, func(_ context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return &mcpsdk.ReadResourceResult{Contents: []*mcpsdk.ResourceContents{
			{URI: req.Params.URI, MIMEType: markdownMIME, Text: databasesGuide},
		}}, nil
	})
}

func readTemplateResource(ctx context.Context, deps *Deps, uri string) (*mcpsdk.ReadResourceResult, error) {
	// The SDK's own error, as it is: it carries the protocol's code for a
	// resource that is not there.
	notFound := mcpsdk.ResourceNotFoundError(uri)
	name, err := url.PathUnescape(strings.TrimPrefix(uri, templateURIPrefix))
	if err != nil || name == "" || strings.Contains(name, "/") {
		return nil, notFound //nolint:wrapcheck
	}
	var resp struct {
		Data struct {
			Title       string `json:"title"`
			Tagline     string `json:"tagline"`
			Description string `json:"description"`
		} `json:"data"`
	}
	call := &Call{deps: deps}
	if err = call.Get(ctx, "/app-templates/"+url.PathEscape(name), nil, &resp); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			return nil, notFound //nolint:wrapcheck
		}
		return nil, err
	}
	text := "# " + resp.Data.Title + "\n\n" + resp.Data.Tagline + "\n\n" + resp.Data.Description
	return &mcpsdk.ReadResourceResult{Contents: []*mcpsdk.ResourceContents{
		{URI: uri, MIMEType: markdownMIME, Text: text},
	}}, nil
}

func addPrompts(s *mcpsdk.Server, a access) {
	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "debug_app",
		Title:       "Why is this app not working?",
		Description: "Looks at an app's containers, deployments and logs, and says what is wrong.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: argProject, Description: "the project's key or name", Required: true},
			{Name: argEnv, Description: "the env's name, such as prod", Required: true},
			{Name: argApp, Description: "the app's key or name", Required: true},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		args := req.Params.Arguments
		return userPrompt("Find out why the app " + args[argApp] + " of " + args[argProject] + ", env " + args[argEnv] +
			", is not working as it should.\n\n" +
			"1. get_app_status: are its containers running, and if not, what error stopped them?\n" +
			"2. list_app_deployments with pageLimit 5: how did its last deployments end? For one that " +
			"failed, get_app_deployment_logs.\n" +
			"3. get_app_logs with grep /error|exception|fatal|panic/ and duration 1h; then the last " +
			"100 lines without grep, for what happened just before.\n" +
			"4. If the cause is outside the app - a node down, memory - list_attention and list_nodes.\n\n" +
			"Then say what is wrong, quote the lines that show it, and what to change. " + debugEnding(a)), nil
	})

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "install_app",
		Title:       "Install something from the app store",
		Description: "Finds templates for what is asked, compares them, and checks the chosen one's install.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "what", Description: "what to install, such as a database or a chat server", Required: true},
			{Name: argProject, Description: "the project to install into, to check the install there"},
			{Name: argEnv, Description: "the env to install into"},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		args := req.Params.Arguments
		text := "I want to install " + args["what"] + " on HivePaaS.\n\n" +
			"1. search_templates for it, and compare at most three candidates: what each installs, " +
			"the apps it brings along, and whether it needs the Docker API or extra capabilities.\n" +
			"2. get_template of the one you recommend, and list the parameters I must decide.\n"
		if args[argProject] != "" && args[argEnv] != "" {
			text += "3. preflight_install it into " + args[argProject] + ", env " + args[argEnv] +
				", with the defaults, and tell me what would stop it.\n"
		}
		text += installEnding(a, args[argProject] != "" && args[argEnv] != "")
		return userPrompt(text), nil
	})

	databasePrompts(s, a)
}

// databasePrompts are the prompts about databases: connecting an app to one,
// running one from its image, reaching one from outside.
func databasePrompts(s *mcpsdk.Server, a access) {
	appArgs := func(appDesc string) []*mcpsdk.PromptArgument {
		return []*mcpsdk.PromptArgument{
			{Name: argProject, Description: "the project's key or name", Required: true},
			{Name: argEnv, Description: "the env's name, such as prod", Required: true},
			{Name: argApp, Description: appDesc, Required: true},
		}
	}
	where := func(args map[string]string) string {
		return args[argApp] + " of " + args[argProject] + ", env " + args[argEnv]
	}

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "connect_app_to_database",
		Title:       "Connect an app to a database",
		Description: "Adds the env vars that connect an app to a database or cache of its env, as references.",
		Arguments: append(appArgs("the app to connect, by key or name"),
			&mcpsdk.PromptArgument{Name: "database", Description: "the database or cache, by key or name; " +
				"left out, choose with me"}),
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		args := req.Params.Arguments
		target := "the database I name - list_env_link_targets, and ask me which"
		if args["database"] != "" {
			target = args["database"]
		}
		return userPrompt("Connect the app " + where(args) + " to " + target + ". The resource " +
			databasesURI + " says how HivePaaS does it.\n\n" +
			"1. get_env_link_suggestions for the app and that target. Take the recommended group unless " +
			"the app's own docs or code want other names; keep each value as the reference it is.\n" +
			"2. get_app_settings, kind env-vars: say which of the names it has already, and do not " +
			"replace one without asking me.\n" +
			"3. Say any warning the suggestions carry, such as a target with no port.\n\n" +
			changeEnding(a, NeedWrite, "plan_update_app_settings, kind env-vars, with the variables "+
				"added to runtimeEnvVars and the others kept")), nil
	})

	s.AddPrompt(&mcpsdk.Prompt{
		Name:  "run_database_from_image",
		Title: "Set up a database run from its own image",
		Description: "Gives an app running a database or cache image the env vars or command it needs, " +
			"from its App Kind's credentials.",
		Arguments: append(appArgs("the app running the image, by key or name"),
			&mcpsdk.PromptArgument{Name: "engine", Description: "the engine, such as postgres; " +
				"left out, App Kind's"}),
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		args := req.Params.Arguments
		engine := ""
		if args["engine"] != "" {
			engine = " with engine " + args["engine"]
		}
		return userPrompt("The app " + where(args) + " runs a database or cache from its own image. Set " +
			"it up from its App Kind's credentials; the resource " + databasesURI + " says how.\n\n" +
			"1. get_app_settings, kind kind: its category, engine and which credentials are set (they " +
			"are masked). If the category is not database or cache, or a credential is missing, say what " +
			"to set in App Kind - a password I choose, never one you make up and keep to yourself.\n" +
			"2. get_env_self_suggestions" + engine + ": the variables, or for Redis and Valkey the " +
			"command, and its warnings.\n" +
			"3. If the engine reads them only on its first start (initOnly) and the app has run before, " +
			"tell me they will not change the password of data it has already.\n\n" +
			changeEnding(a, NeedWrite, "plan_update_app_settings - kind env-vars for the variables, "+
				"the others kept; kind deployment for a command")), nil
	})

	s.AddPrompt(&mcpsdk.Prompt{
		Name:  "expose_database",
		Title: "Reach a database from outside",
		Description: "Gives a database or cache a TCP domain, or for MySQL and MariaDB a published port, " +
			"so a client outside the cluster can reach it.",
		Arguments: append(appArgs("the database or cache, by key or name"),
			&mcpsdk.PromptArgument{Name: paramDomain, Description: "the domain to reach it at, such as " +
				"db.example.com; left out, ask me"}),
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		args := req.Params.Arguments
		domain := "a domain - ask me which, and whether its DNS points at the servers"
		if args[paramDomain] != "" {
			domain = args[paramDomain]
		}
		return userPrompt("Make the database " + where(args) + " reachable from outside the cluster, at " +
			domain + ". The resource " + databasesURI + " says how; do not change its App Kind for it.\n\n" +
			"1. get_app_settings, kind kind: its engine. For MySQL or MariaDB a domain cannot work - " +
			"say so, and offer publishing its port in the network settings instead.\n" +
			"2. get_app_settings, kind routing: its port and the domains it has. Keep them.\n" +
			"3. Ask me whether Traefik should end TLS (the usual) or pass it through to the database, " +
			"which then needs its own certificate mounted.\n\n" +
			changeEnding(a, NeedWrite, "plan_update_app_settings, kind routing: exposePublicly true, and "+
				"the domain added with protocol tcp and the database's port as containerPort, the others "+
				"kept") +
			" Then tell me to open the port in the firewall, and give me a command to connect with, " +
			"over TLS."), nil
	})
}

// changeEnding is how a task that changes settings ends: planned and applied
// once the person agrees, or said, when this key or server may not.
func changeEnding(a access, need Need, plan string) string {
	switch {
	case a.serves(need):
		return "Then " + plan + ". Show me the plan, and apply it only once I agree."
	case !a.changesAllowed:
		return "Change nothing: changes are off on this server. Say what to change, and that the " +
			"dashboard can do it."
	}
	return "Change nothing: this key may not. Say what to change - it takes a key with Write - and " +
		"that the dashboard can do it."
}

// debugEnding is what a diagnosis ends with: a change the tools can make is
// planned and shown, and one they cannot is described, with what would allow it.
func debugEnding(a access) string {
	switch {
	case !a.changesAllowed:
		return "Change nothing: these tools only read."
	case !a.mayChange():
		return "Change nothing: this key may only read. If a change would fix it, say which: restarting " +
			"or redeploying needs a key with Execute; changing configuration, one with Write."
	}
	return "If a restart, a redeploy or a configuration change would fix it, plan it with the plan_* " +
		"tool, show me the plan, and apply it only once I agree. If the change needs access this key " +
		"does not have, say which."
}

// installEnding is what an install ends with: planned and applied once the
// person agrees, or done in the dashboard.
func installEnding(a access, placed bool) string {
	writable := a.serves(NeedWrite)
	switch {
	case writable && placed:
		return "4. Then plan_install_app with what I decided, show me the plan, and apply it only once I agree."
	case writable:
		return "\nAsk me which project and env to install into, then plan_install_app, show me the plan, " +
			"and apply it only once I agree."
	}
	if a.changesAllowed {
		return "\nThis key cannot install: installing needs a key with Write. Say so, and that installing " +
			"can also be done in the dashboard."
	}
	return "\nInstalling itself is done in the dashboard; say where."
}

func userPrompt(text string) *mcpsdk.GetPromptResult {
	return &mcpsdk.GetPromptResult{Messages: []*mcpsdk.PromptMessage{
		{Role: "user", Content: &mcpsdk.TextContent{Text: text}},
	}}
}
