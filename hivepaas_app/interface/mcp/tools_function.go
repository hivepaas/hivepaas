package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// functionFileInput is one file of a function's code.
type functionFileInput struct {
	Path    string `json:"path" jsonschema:"the file's path in the function, such as index.ts, main.py or lib/db.py"`
	Content string `json:"content" jsonschema:"the file's whole content"`
}

// planFile is a file as a plan shows it: its path and size, not its content,
// which the model sent and the person can ask for.
type planFile struct {
	Path  string `json:"path"`
	Bytes int    `json:"bytes"`
}

func planFiles(files []*appsettingsdto.FunctionFileReq) []planFile {
	out := make([]planFile, 0, len(files))
	for _, f := range files {
		out = append(out, planFile{Path: f.Path, Bytes: len(f.Content)})
	}
	return out
}

func inlineCode(files []functionFileInput) *appsettingsdto.FunctionInlineCodeReq {
	code := &appsettingsdto.FunctionInlineCodeReq{Files: make([]*appsettingsdto.FunctionFileReq, 0, len(files))}
	for _, f := range files {
		code.Files = append(code.Files, &appsettingsdto.FunctionFileReq{Path: f.Path, Content: f.Content})
	}
	return code
}

// validationProblem words the API's own checks of a request, run before the
// plan is made, as one input error: "source.entrypoint.file: ..." per line.
func validationProblem(what string, validators []vld.Validator) error {
	errs := vld.Validate(validators...)
	if len(errs) == 0 {
		return nil
	}
	lines := make([]string, 0, len(errs))
	for _, e := range errs {
		field := ""
		if f := e.Field(); f != nil {
			field = f.PathString(true, ".") + ": "
		}
		key, _ := e.CustomKey().(string)
		if key == "" {
			key = e.Error()
		}
		lines = append(lines, field+key)
	}
	sort.Strings(lines)
	return &InputError{Message: what + " would be refused:\n" + strings.Join(lines, "\n")}
}

// ---- plan_create_function ----

type createFunctionInput struct {
	Project string              `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string              `json:"env" jsonschema:"the env's name, such as production"`
	Name    string              `json:"name" jsonschema:"the function's name, up to 100 characters"`
	Runtime string              `json:"runtime" jsonschema:"the runtime: node24, bun1, python313 or go127"`
	Files   []functionFileInput `json:"files" jsonschema:"every file of the function's code, its library manifest too"`
	//nolint:lll // a schema description is one line
	Entrypoint string `json:"entrypoint,omitempty" jsonschema:"the handler's file - for go127 its package's directory; the runtime's default when not given"`
	Handler    string `json:"handler,omitempty" jsonschema:"the handler's name; the runtime's default when not given"`
	//nolint:lll // a schema description is one line
	Domain string   `json:"domain,omitempty" jsonschema:"a domain to route it at over HTTPS from its first deployment; none, and it is reached inside its project"`
	Note   string   `json:"note,omitempty" jsonschema:"a note on the function, shown with it"`
	Tags   []string `json:"tags,omitempty" jsonschema:"tags to file the function under"`
}

type createFunctionPlan struct {
	Project    string     `json:"project"`
	Env        string     `json:"env"`
	Name       string     `json:"name"`
	Runtime    string     `json:"runtime"`
	Entrypoint string     `json:"entrypoint"`
	Files      []planFile `json:"files"`
	Domain     string     `json:"domain,omitempty"`
	Effect     string     `json:"effect"`
}

func planCreateFunctionTool() Tool {
	return planTool("plan_create_function", "Plan creating a function",
		"Plans POST /projects/{project}/{env}/apps/function: a function - an app whose deployment is code and "+
			"a runtime ("+statusValues(base.AllFunctionRuntimes)+") - created and its first deployment queued "+
			"at once: built on the runtime's image from the files given, then started. Every file goes in "+
			"files, a package.json, requirements.txt or go.mod among them for its libraries. The API's own "+
			"checks - the entrypoint's extension, file paths, sizes - are run before the plan is made. The "+
			"resource "+functionsURI+" says how a handler is written for each runtime. Nothing is created "+
			"until apply_plan.",
		NeedWrite, &applier{follow: "list_app_deployments with the new id follows the build; get_app_status " +
			"shows it start; plan_test_run_function calls it."},
		func(ctx context.Context, call *Call, in createFunctionInput) (createFunctionPlan, *storedPlan, error) {
			name := strings.TrimSpace(in.Name)
			if name == "" {
				return createFunctionPlan{}, nil, &InputError{Message: "name is required"}
			}
			if len(in.Files) == 0 {
				return createFunctionPlan{}, nil, &InputError{Message: "files is required: the function's code"}
			}
			ref, err := resolveEnv(ctx, call, in.Project, in.Env)
			if err != nil {
				return createFunctionPlan{}, nil, err
			}
			apps, err := listApps(ctx, call, ref)
			if err != nil {
				return createFunctionPlan{}, nil, err
			}
			if taken := appNamed(apps, name); taken != nil {
				return createFunctionPlan{}, nil, &InputError{Message: fmt.Sprintf(
					"%s/%s has an app named %s already (%s): choose another name", ref.ProjectKey, ref.Env, name,
					taken.Key)}
			}

			source := &appsettingsdto.DeploymentFunctionSourceReq{
				Runtime:    base.FunctionRuntime(in.Runtime),
				Entrypoint: appsettingsdto.FunctionEntrypointReq{File: in.Entrypoint, Handler: in.Handler},
				Code:       appsettingsdto.FunctionCodeReq{Inline: inlineCode(in.Files)},
			}
			source.Normalize()
			if err = validationProblem("The function", source.Validate("source")); err != nil {
				return createFunctionPlan{}, nil, err
			}
			domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(in.Domain)), ".")

			body := &appdto.CreateFunctionReq{
				AppBaseReq: &appdto.AppBaseReq{Name: name, Status: base.AppStatusActive, Note: in.Note,
					Tags: in.Tags},
				Source: source,
				Domain: domain,
			}
			raw, err := json.Marshal(body)
			if err != nil {
				return createFunctionPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}

			reach := "it is reached inside its project, by its service's name, on port 8080"
			if domain != "" {
				reach = "it is routed at https://" + domain + " from its first deployment"
			}
			out := createFunctionPlan{
				Project: ref.ProjectKey, Env: ref.Env, Name: name, Runtime: string(source.Runtime),
				Entrypoint: source.Entrypoint.File + " / " + source.Entrypoint.Handler,
				Files:      planFiles(source.Code.Inline.Files), Domain: domain,
				Effect: "the function is created and deployed: built on the " + string(source.Runtime) +
					" image, its libraries installed, then started; " + reach,
			}
			return out, &storedPlan{Method: http.MethodPost, Path: ref.path("/apps/function"), Body: raw,
				Summary: fmt.Sprintf("create the function %s (%s) in %s/%s", name, source.Runtime, ref.ProjectKey,
					ref.Env)}, nil
		})
}

// appNamed is the app of the env with this name or key, whatever the case;
// nil when there is none.
func appNamed(apps []listedApp, name string) *listedApp {
	for i := range apps {
		if strings.EqualFold(apps[i].Name, name) || strings.EqualFold(apps[i].Key, name) {
			return &apps[i]
		}
	}
	return nil
}

// ---- plan_test_run_function ----

type testRunInput struct {
	Project string `json:"project" jsonschema:"the project's key, name or id; list_projects lists them"`
	Env     string `json:"env" jsonschema:"the env's name, such as production"`
	App     string `json:"app" jsonschema:"the function, by key, name or id"`
	//nolint:lll // a schema description is one line
	Files   []functionFileInput `json:"files,omitempty" jsonschema:"code to run instead of the function's saved code - every file of it; the saved code when not given"`
	Method  string              `json:"method,omitempty" jsonschema:"the request's method; GET when not given"`
	Path    string              `json:"path,omitempty" jsonschema:"the request's path, no query; / when not given"`
	Query   map[string]string   `json:"query,omitempty" jsonschema:"the request's query, a value per name"`
	Headers map[string]string   `json:"headers,omitempty" jsonschema:"the request's headers, a value per name"`
	Body    string              `json:"body,omitempty" jsonschema:"the request's body, as text"`
}

type testRunPlan struct {
	Project string     `json:"project"`
	Env     string     `json:"env"`
	App     string     `json:"app"`
	Code    string     `json:"code"`
	Files   []planFile `json:"files"`
	Request string     `json:"request"`
	Effect  string     `json:"effect"`
}

// testRunAnswer is a test run's answer as a model reads it: the body as text
// when it is text.
type testRunAnswer struct {
	Outcome        string              `json:"outcome"`
	Status         int                 `json:"status,omitempty"`
	Headers        map[string][]string `json:"headers,omitempty"`
	Body           string              `json:"body,omitempty"`
	BodyBase64     []byte              `json:"bodyBase64,omitempty"`
	BodyTruncated  bool                `json:"bodyTruncated,omitempty"`
	DurationMs     float64             `json:"durationMs,omitempty"`
	Logs           string              `json:"logs"`
	LogsTruncated  bool                `json:"logsTruncated,omitempty"`
	Error          string              `json:"error,omitempty"`
	ExitCode       int64               `json:"exitCode"`
	LibrariesBuilt bool                `json:"librariesBuilt"`
	LibrariesLog   string              `json:"librariesLog,omitempty"`
	LockFiles      []string            `json:"lockFiles,omitempty"`
}

func testRunResult(data json.RawMessage) (any, error) {
	var resp appdto.TestRunFunctionDataResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("mcp: reading a test run: %w", err)
	}
	out := testRunAnswer{
		Outcome: resp.Outcome, Status: resp.Status, Headers: resp.Headers, BodyTruncated: resp.BodyTruncated,
		DurationMs: resp.DurationMs, Logs: resp.Logs, LogsTruncated: resp.LogsTruncated, Error: resp.Error,
		ExitCode: resp.ExitCode, LibrariesBuilt: resp.LibrariesBuilt, LibrariesLog: resp.LibrariesLog,
	}
	if utf8.Valid(resp.Body) {
		out.Body = string(resp.Body)
	} else {
		out.BodyBase64 = resp.Body
	}
	for _, f := range resp.LockFiles {
		out.LockFiles = append(out.LockFiles, f.Path)
	}
	return out, nil
}

func planTestRunFunctionTool() Tool {
	return planTool("plan_test_run_function", "Plan a function's test run",
		"Plans POST /projects/{project}/{env}/apps/{app}/function/test-run: the function's code - its saved "+
			"code, or the files given, not yet saved - called once with a request, in a throwaway container "+
			"on a build node, with the function's variables and secrets, as the dashboard's Test does. Nothing "+
			"is saved or deployed. Applied, it answers the response - status, headers, body - the call's log, "+
			"the runtime's exit code and error, and, when its libraries were installed, the lock files made. "+
			"What the code prints is answered as it is, a secret it prints included. The API takes it as a "+
			"change to the function, so it needs Write.",
		NeedWrite, &applier{result: testRunResult,
			follow: "plan_update_app_settings, kind deployment, saves code that worked; plan_redeploy_app " +
				"deploys it."},
		func(ctx context.Context, call *Call, in testRunInput) (testRunPlan, *storedPlan, error) {
			app, err := resolveApp(ctx, call, in.Project, in.Env, in.App)
			if err != nil {
				return testRunPlan{}, nil, err
			}

			code := "the files given"
			var files *appsettingsdto.FunctionInlineCodeReq
			if len(in.Files) > 0 {
				files = inlineCode(in.Files)
			} else {
				code = "the function's saved code"
				if files, err = savedFunctionCode(ctx, call, app); err != nil {
					return testRunPlan{}, nil, err
				}
			}

			req := &appdto.TestRunFunctionReq{
				Code: files,
				Request: &appdto.TestRunRequestReq{Method: in.Method, Path: in.Path, Query: multi(in.Query),
					Headers: multi(in.Headers), Body: in.Body},
			}
			// As the endpoint reads it: GET / unless said, header names in lower case.
			if err = req.ModifyRequest(); err != nil {
				return testRunPlan{}, nil, fmt.Errorf("mcp: reading a test run: %w", err)
			}
			// The code is what a model gets wrong; the request the endpoint checks.
			if err = validationProblem("The code", req.Code.Validate("code")); err != nil {
				return testRunPlan{}, nil, err
			}
			raw, err := json.Marshal(req)
			if err != nil {
				return testRunPlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}

			out := testRunPlan{
				Project: app.ProjectKey, Env: app.Env, App: app.AppKey, Code: code,
				Files:   planFiles(files.Files),
				Request: req.Request.Method + " " + req.Request.Path,
				Effect: "the code runs once, in a container thrown away after, with " + app.AppKey +
					"'s variables and secrets; nothing is saved or deployed",
			}
			return out, &storedPlan{Method: http.MethodPost, Path: app.path("/function/test-run"), Body: raw,
				Summary: fmt.Sprintf("test-run %s's %s with %s", app.AppKey, code, out.Request)}, nil
		})
}

// savedFunctionCode is a function's inline code as saved in its deployment
// settings; an error a model can act on when the app is no function, or its
// code is in a repository.
func savedFunctionCode(ctx context.Context, call *Call, app *appRef) (*appsettingsdto.FunctionInlineCodeReq, error) {
	var resp struct {
		Data appsettingsdto.DeploymentSettingsResp `json:"data"`
	}
	if err := call.Get(ctx, app.path("/deployment-settings"), nil, &resp); err != nil {
		return nil, err
	}
	source := resp.Data.FunctionSource
	if resp.Data.ActiveMethod != base.DeploymentMethodFunction || source == nil {
		return nil, &InputError{Message: app.AppKey + " is not a function: it deploys from " +
			string(resp.Data.ActiveMethod)}
	}
	if source.Code == nil || source.Code.Inline == nil || len(source.Code.Inline.Files) == 0 {
		return nil, &InputError{Message: app.AppKey + "'s code is in a repository: give the files to run"}
	}
	code := &appsettingsdto.FunctionInlineCodeReq{}
	for _, f := range source.Code.Inline.Files {
		code.Files = append(code.Files, &appsettingsdto.FunctionFileReq{Path: f.Path, Content: f.Content})
	}
	return code, nil
}

// multi is a value per name as the API takes it, a list per name.
func multi(values map[string]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string][]string, len(values))
	for k, v := range values {
		out[k] = []string{v}
	}
	return out
}
