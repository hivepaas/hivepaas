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
	markdownMIME      = "text/markdown"
)

// dockerAPIGuide is the dashboard's permissions guide, as text. Kept beside
// pkg/dockerproxy and changed with it, as the dashboard's copy is.
//
//go:embed docs/docker-api.md
var dockerAPIGuide string

func addResources(s *mcpsdk.Server, deps *Deps) {
	s.AddResourceTemplate(&mcpsdk.ResourceTemplate{
		URITemplate: templateURIPrefix + "{name}",
		Name:        "template",
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

func addPrompts(s *mcpsdk.Server) {
	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "debug_app",
		Title:       "Why is this app not working?",
		Description: "Looks at an app's containers, deployments and logs, and says what is wrong.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "project", Description: "the project's key or name", Required: true},
			{Name: "env", Description: "the env's name, such as prod", Required: true},
			{Name: "app", Description: "the app's key or name", Required: true},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		a := req.Params.Arguments
		return userPrompt("Find out why the app " + a["app"] + " of " + a["project"] + ", env " + a["env"] +
			", is not working as it should.\n\n" +
			"1. get_app_status: are its containers running, and if not, what error stopped them?\n" +
			"2. get_app: how did its last deployments end?\n" +
			"3. get_app_logs with grep /error|exception|fatal|panic/ over the last hour; then the last " +
			"100 lines without grep, for what happened just before.\n" +
			"4. If the cause is outside the app - a node down, memory - list_attention and list_nodes.\n\n" +
			"Then say what is wrong, quote the lines that show it, and what to change. Change nothing: " +
			"these tools only read."), nil
	})

	s.AddPrompt(&mcpsdk.Prompt{
		Name:        "install_app",
		Title:       "Install something from the app store",
		Description: "Finds templates for what is asked, compares them, and checks the chosen one's install.",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "what", Description: "what to install, such as a database or a chat server", Required: true},
			{Name: "project", Description: "the project to install into, to check the install there"},
			{Name: "env", Description: "the env to install into"},
		},
	}, func(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		a := req.Params.Arguments
		text := "I want to install " + a["what"] + " on HivePaaS.\n\n" +
			"1. search_templates for it, and compare at most three candidates: what each installs, " +
			"the apps it brings along, and whether it needs the Docker API or extra capabilities.\n" +
			"2. get_template of the one you recommend, and list the parameters I must decide.\n"
		if a["project"] != "" && a["env"] != "" {
			text += "3. preflight_install it into " + a["project"] + ", env " + a["env"] +
				", with the defaults, and tell me what would stop it.\n"
		}
		text += "\nInstalling itself is done in the dashboard; say where."
		return userPrompt(text), nil
	})
}

func userPrompt(text string) *mcpsdk.GetPromptResult {
	return &mcpsdk.GetPromptResult{Messages: []*mcpsdk.PromptMessage{
		{Role: "user", Content: &mcpsdk.TextContent{Text: text}},
	}}
}
