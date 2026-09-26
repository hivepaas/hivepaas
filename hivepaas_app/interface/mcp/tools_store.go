package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// ---- search_templates ----

// maxTemplates is the most templates search_templates answers.
const maxTemplates = 50

// maxDescription is the most of a template's description an answer quotes.
const maxDescription = 8 << 10

type searchTemplatesInput struct {
	Search   string `json:"search,omitempty" jsonschema:"text in a template's name, title, tagline, tags or aliases"`
	Category string `json:"category,omitempty" jsonschema:"a category such as databases or databases/sql"`
}

type templateItem struct {
	Name                 string                                   `json:"name"`
	Title                string                                   `json:"title"`
	Tagline              string                                   `json:"tagline"`
	Categories           []string                                 `json:"categories,omitempty"`
	Versions             []*apptemplatedto.AppTemplateVersionResp `json:"versions,omitempty"`
	Variants             []string                                 `json:"variants,omitempty"`
	Dependencies         []string                                 `json:"dependencies,omitempty"`
	Components           []string                                 `json:"components,omitempty"`
	RequiresDockerAPI    bool                                     `json:"requiresDockerApi,omitempty"`
	RequiresCapabilities bool                                     `json:"requiresCapabilities,omitempty"`
	Compatible           bool                                     `json:"compatible"`
}

type templateList struct {
	Templates []templateItem `json:"templates"`
	Truncated int            `json:"truncated,omitempty"`
}

func (l *templateList) shrink() bool {
	return shrinkList(&l.Templates, &l.Truncated)
}

func searchTemplatesTool() Tool {
	return readTool("search_templates", "Search the app store",
		"Searches the templates of the app store: what each installs, its versions, the apps it "+
			"brings along as dependencies, and whether it needs the Docker API or extra capabilities. "+
			"A template that is not compatible needs a newer HivePaaS.",
		func(ctx context.Context, call *Call, in searchTemplatesInput) (templateList, error) {
			query := url.Values{paramPageLimit: {strconv.Itoa(maxTemplates)}}
			if s := strings.TrimSpace(in.Search); s != "" {
				query.Set("search", s)
			}
			if c := strings.TrimSpace(in.Category); c != "" {
				query.Set("category", c)
			}
			var resp apptemplatedto.ListAppTemplatesResp
			if err := call.Get(ctx, "/app-templates", query, &resp); err != nil {
				return templateList{}, err
			}
			out := templateList{Templates: make([]templateItem, 0, len(resp.Data))}
			for _, t := range resp.Data {
				if t != nil {
					out.Templates = append(out.Templates, makeTemplateItem(t))
				}
			}
			return out, nil
		})
}

func makeTemplateItem(t *apptemplatedto.AppTemplateSummaryResp) templateItem {
	item := templateItem{Name: t.Name, Title: t.Title, Tagline: t.Tagline, Categories: t.Categories,
		Versions: t.Versions, RequiresDockerAPI: t.RequiresDockerAPI,
		RequiresCapabilities: t.RequiresCapabilities, Compatible: t.Compatible}
	for _, d := range t.Dependencies {
		if d != nil {
			item.Dependencies = append(item.Dependencies, d.Template)
		}
	}
	for _, c := range t.Components {
		if c != nil {
			item.Components = append(item.Components, c.Name)
		}
	}
	for _, v := range t.Variants {
		if v != nil {
			item.Variants = append(item.Variants, v.Name)
		}
	}
	return item
}

// ---- get_template ----

type templateInput struct {
	Template string `json:"template" jsonschema:"the template's name, from search_templates"`
}

// templateDetail is a template as the store's endpoint answers it, less what
// only the dashboard uses: its own, public text.
type templateDetail map[string]any

// templateOnlyForTheDashboard are the fields of a template a model has no use
// for: where the store read it from, and its icon.
var templateOnlyForTheDashboard = []string{"source", "revision", "iconUrl"}

func getTemplateTool() Tool {
	return readTool("get_template", "Get a template",
		"Shows one template of the app store in full: its description, the parameters an install "+
			"asks for (a secret one is generated when left empty), its dependencies and components, "+
			"and what it is granted - Docker API access, capabilities, published ports.",
		func(ctx context.Context, call *Call, in templateInput) (templateDetail, error) {
			name := strings.TrimSpace(in.Template)
			if name == "" {
				return templateDetail{}, &InputError{Message: "template is required; search_templates lists them"}
			}
			var resp struct {
				Data json.RawMessage `json:"data"`
			}
			if err := call.Get(ctx, "/app-templates/"+url.PathEscape(name), nil, &resp); err != nil {
				return templateDetail{}, err
			}
			out, err := through(resp.Data, &apptemplatedto.AppTemplateResp{})
			if err != nil {
				return templateDetail{}, fmt.Errorf("mcp: reading the template: %w", err)
			}
			for _, field := range templateOnlyForTheDashboard {
				delete(out, field)
			}
			if description, ok := out["description"].(string); ok {
				out["description"] = cutText(description, maxDescription, "")
			}
			return out, nil
		})
}

// ---- preflight_install ----

type preflightInput struct {
	Project          string                    `json:"project" jsonschema:"the project's key, name or id"`
	Env              string                    `json:"env" jsonschema:"the env's name, such as prod"`
	Template         string                    `json:"template" jsonschema:"the template's name"`
	Name             string                    `json:"name" jsonschema:"the name the new app would have"`
	Version          string                    `json:"version,omitempty" jsonschema:"a version; the default when empty"`
	Variant          string                    `json:"variant,omitempty" jsonschema:"a variant; the default when empty"`
	Params           map[string]any            `json:"params,omitempty" jsonschema:"the template's parameters by name"`
	DependencyParams map[string]map[string]any `json:"dependencyParams,omitempty" jsonschema:"by dependency"`
}

// body is the create endpoint's own request: preflight asks about it, and an
// install sends it.
func (in preflightInput) body() *apptemplatedto.CreateAppFromTemplateReq {
	return &apptemplatedto.CreateAppFromTemplateReq{Name: in.Name, Template: in.Template, Version: in.Version,
		Variant: in.Variant, Params: in.Params, DependencyParams: in.DependencyParams}
}

// forAudit keeps the parameters' names: which of them are secret is the
// template's to say, and any may be.
func (in preflightInput) forAudit() any {
	redacted := in
	redacted.Params = redactValues(in.Params)
	redacted.DependencyParams = make(map[string]map[string]any, len(in.DependencyParams))
	for dep, params := range in.DependencyParams {
		redacted.DependencyParams[dep] = redactValues(params)
	}
	return redacted
}

func redactValues(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for name := range params {
		out[name] = "(given)"
	}
	return out
}

type preflightAnswer struct {
	// Issues are what would refuse the install; none means it would go ahead.
	Issues []*apptemplatedto.PreflightIssueResult `json:"issues"`
	// Storage is where an earlier install left data the new apps would find.
	Storage []*apptemplatedto.PreflightStorageResult `json:"storage,omitempty"`
	// StorageUnchecked is where the check could not look.
	StorageUnchecked []*apptemplatedto.PreflightStorageResult `json:"storageUnchecked,omitempty"`
	// Apps are what the install would create.
	Apps []*apptemplatedto.PreflightPlannedApp `json:"apps,omitempty"`
}

func preflightInstallTool() Tool {
	return readTool("preflight_install", "Check an install",
		"Checks, without installing anything, whether a template would install into an env as asked: "+
			"the refusals the install would meet, and data an earlier install left where the new apps "+
			"would keep theirs. It needs a key that may create apps in the env, since it looks into "+
			"the env's volumes.",
		func(ctx context.Context, call *Call, in preflightInput) (preflightAnswer, error) {
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return preflightAnswer{}, err
			}
			result, err := preflight(ctx, call, ref, in.body())
			if err != nil {
				return preflightAnswer{}, err
			}
			return preflightAnswer{Issues: result.Issues, Apps: result.Apps, Storage: result.Storage,
				StorageUnchecked: result.StorageUnchecked}, nil
		})
}

// preflight asks the create endpoint's preflight about a request, as the caller.
func preflight(ctx context.Context, call *Call, ref *envRef, body *apptemplatedto.CreateAppFromTemplateReq) (
	*apptemplatedto.PreflightAppResult, error) {
	var resp apptemplatedto.PreflightAppFromTemplateResp
	if err := call.Post(ctx, ref.path("/apps/from-template/preflight"), body, &resp); err != nil {
		return nil, err
	}
	result := resp.Data
	if result == nil {
		result = &apptemplatedto.PreflightAppResult{}
	}
	if result.Issues == nil {
		result.Issues = []*apptemplatedto.PreflightIssueResult{}
	}
	return result, nil
}
