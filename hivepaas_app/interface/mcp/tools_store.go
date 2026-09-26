package mcp

import (
	"context"
	"net/url"
	"strconv"
	"strings"
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

type templateVersion struct {
	Name    string `json:"name"`
	Release string `json:"release,omitempty"`
	Default bool   `json:"default,omitempty"`
}

type templateItem struct {
	Name                 string            `json:"name"`
	Title                string            `json:"title"`
	Tagline              string            `json:"tagline"`
	Categories           []string          `json:"categories,omitempty"`
	Versions             []templateVersion `json:"versions,omitempty"`
	Variants             []string          `json:"variants,omitempty"`
	Dependencies         []string          `json:"dependencies,omitempty"`
	Components           []string          `json:"components,omitempty"`
	RequiresDockerAPI    bool              `json:"requiresDockerApi,omitempty"`
	RequiresCapabilities bool              `json:"requiresCapabilities,omitempty"`
	Compatible           bool              `json:"compatible"`
}

type templateList struct {
	Templates []templateItem `json:"templates"`
	Truncated int            `json:"truncated,omitempty"`
}

func (l *templateList) shrink() bool {
	return shrinkList(&l.Templates, &l.Truncated)
}

type apiTemplateSummary struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Tagline      string   `json:"tagline"`
	Categories   []string `json:"categories"`
	Dependencies []struct {
		Template string `json:"template"`
	} `json:"dependencies"`
	Components []struct {
		Name string `json:"name"`
	} `json:"components"`
	Variants []struct {
		Name string `json:"name"`
	} `json:"variants"`
	Versions             []templateVersion `json:"versions"`
	RequiresCapabilities bool              `json:"requiresCapabilities"`
	RequiresDockerAPI    bool              `json:"requiresDockerApi"`
	Compatible           bool              `json:"compatible"`
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
			var resp struct {
				Data []apiTemplateSummary `json:"data"`
			}
			if err := call.Get(ctx, "/app-templates", query, &resp); err != nil {
				return templateList{}, err
			}
			out := templateList{Templates: make([]templateItem, 0, len(resp.Data))}
			for _, t := range resp.Data {
				out.Templates = append(out.Templates, makeTemplateItem(&t))
			}
			return out, nil
		})
}

func makeTemplateItem(t *apiTemplateSummary) templateItem {
	item := templateItem{Name: t.Name, Title: t.Title, Tagline: t.Tagline, Categories: t.Categories,
		Versions: t.Versions, RequiresDockerAPI: t.RequiresDockerAPI,
		RequiresCapabilities: t.RequiresCapabilities, Compatible: t.Compatible}
	for _, d := range t.Dependencies {
		item.Dependencies = append(item.Dependencies, d.Template)
	}
	for _, c := range t.Components {
		item.Components = append(item.Components, c.Name)
	}
	for _, v := range t.Variants {
		item.Variants = append(item.Variants, v.Name)
	}
	return item
}

// ---- get_template ----

type templateInput struct {
	Template string `json:"template" jsonschema:"the template's name, from search_templates"`
}

// templateDetail is a template as the store shows it. The nested parts are
// passed on as the API words them: they are the template's own, public text.
type templateDetail struct {
	Name           string   `json:"name"`
	Title          string   `json:"title"`
	Tagline        string   `json:"tagline"`
	Description    string   `json:"description"`
	Categories     []string `json:"categories,omitempty"`
	License        string   `json:"license,omitempty"`
	Compatible     bool     `json:"compatible"`
	Links          any      `json:"links,omitempty"`
	Versions       any      `json:"versions,omitempty"`
	Variants       any      `json:"variants,omitempty"`
	Parameters     any      `json:"parameters,omitempty"`
	Dependencies   any      `json:"dependencies,omitempty"`
	Components     any      `json:"components,omitempty"`
	Capabilities   any      `json:"capabilities,omitempty"`
	DockerAPI      any      `json:"dockerApi,omitempty"`
	PublishedPorts any      `json:"publishedPorts,omitempty"`
}

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
				Data templateDetail `json:"data"`
			}
			if err := call.Get(ctx, "/app-templates/"+url.PathEscape(name), nil, &resp); err != nil {
				return templateDetail{}, err
			}
			resp.Data.Description = cutText(resp.Data.Description, maxDescription, "")
			return resp.Data, nil
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

type preflightStorage struct {
	App        string `json:"app"`
	AppKey     string `json:"appKey"`
	IsDatabase bool   `json:"isDatabase,omitempty"`
	Volume     string `json:"volume"`
	Path       string `json:"path"`
}

type preflightIssue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type preflightAnswer struct {
	// Issues are what would refuse the install; none means it would go ahead.
	Issues []preflightIssue `json:"issues"`
	// Storage is where an earlier install left data the new apps would find.
	Storage []preflightStorage `json:"storage,omitempty"`
	// StorageUnchecked is where the check could not look.
	StorageUnchecked []preflightStorage `json:"storageUnchecked,omitempty"`
}

type apiPreflightStorage struct {
	App        string `json:"app"`
	AppKey     string `json:"appKey"`
	IsDatabase bool   `json:"isDatabase"`
	Volume     struct {
		Name string `json:"name"`
	} `json:"volume"`
	Path string `json:"path"`
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
			body := map[string]any{"name": in.Name, "template": in.Template, "version": in.Version,
				"variant": in.Variant, "params": in.Params, "dependencyParams": in.DependencyParams}
			var resp struct {
				Data struct {
					Storage          []apiPreflightStorage `json:"storage"`
					StorageUnchecked []apiPreflightStorage `json:"storageUnchecked"`
					Issues           []preflightIssue      `json:"issues"`
				} `json:"data"`
			}
			if err = call.Post(ctx, ref.path("/apps/from-template/preflight"), body, &resp); err != nil {
				return preflightAnswer{}, err
			}
			out := preflightAnswer{Issues: resp.Data.Issues,
				Storage:          makePreflightStorage(resp.Data.Storage),
				StorageUnchecked: makePreflightStorage(resp.Data.StorageUnchecked)}
			if out.Issues == nil {
				out.Issues = []preflightIssue{}
			}
			return out, nil
		})
}

func makePreflightStorage(items []apiPreflightStorage) []preflightStorage {
	var out []preflightStorage
	for _, s := range items {
		out = append(out, preflightStorage{App: s.App, AppKey: s.AppKey, IsDatabase: s.IsDatabase,
			Volume: s.Volume.Name, Path: s.Path})
	}
	return out
}
