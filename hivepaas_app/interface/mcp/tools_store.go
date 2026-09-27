package mcp

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// The tools that install from the app store send the create endpoint's own
// request: preflight_install asks the endpoint's preflight about it, and
// plan_install_app plans sending it.

// installArgs are the arguments of an install: the env's names, and the
// create endpoint's request body.
type installArgs map[string]any

// installDescs describe the create endpoint's request.
var installDescs = map[string]string{
	argTemplate: descTemplate,
	argName: "the new app's name, up to 100 characters; its key is made from it, and the apps it brings " +
		"along are named after it",
	"version": "one of the template's versions (get_template); its default version when not given",
	"variant": "one of the version's variants (get_template); its default variant when not given",
	"imageTag": "a tag of the version's image to run instead of the one it pins, from " +
		"list_template_image_tags; the pinned tag when not given",
	"params": "the template's parameters by name, each a JSON value of its type - a size is a string such " +
		"as 1GB, and a volume is a volume's id from list_volumes. A parameter not given takes its default; a " +
		"secret one not given is generated",
	"dependencyParams": "the parameters of each app the template brings along, by the dependency's name " +
		"(get_template's dependencies), each as params",
}

// installInput is the input schema of the install tools. resetStorage, which
// deletes what an earlier install left, is the dashboard's to ask for.
var installInput = bodyInput(underEnv, &apptemplatedto.CreateAppFromTemplateReq{}, installDescs,
	[]string{argTemplate, argName}, "resetStorage")

// resolve finds the env, and reads the request body.
func (in installArgs) resolve(ctx context.Context, call *Call) (*envRef, *apptemplatedto.CreateAppFromTemplateReq,
	error) {
	project, _ := in[argProject].(string)
	env, _ := in[argEnv].(string)
	ref, err := resolveEnv(ctx, call, project, env)
	if err != nil {
		return nil, nil, err
	}
	body := &apptemplatedto.CreateAppFromTemplateReq{}
	if err = decodeBody(in, body); err != nil {
		return nil, nil, err
	}
	body.ResetStorage = false
	return ref, body, nil
}

// forAudit keeps the parameters' names: which of them are secret is the
// template's to say, and any may be.
func (in installArgs) forAudit() any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	if params, ok := in["params"].(map[string]any); ok {
		out["params"] = redactValues(params)
	}
	if deps, ok := in["dependencyParams"].(map[string]any); ok {
		redacted := make(map[string]any, len(deps))
		for dep, params := range deps {
			p, _ := params.(map[string]any)
			redacted[dep] = redactValues(p)
		}
		out["dependencyParams"] = redacted
	}
	return out
}

func redactValues(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for name := range params {
		out[name] = redactedValue
	}
	return out
}

// ---- preflight_install ----

func preflightInstallTool() Tool {
	return readToolWith("preflight_install", "Check an install",
		"POST /projects/{project}/{env}/apps/from-template/preflight. Checks, without creating anything, "+
			"what installing a template into an env would do: the apps it would create (data.apps - the "+
			"dependencies first, then the app asked for, each with its name, key and image), the refusals "+
			"the install would meet (data.issues - a name or domain taken, a port held, a permission "+
			"missing; none means it would go ahead), and data an earlier install left where the new apps "+
			"would keep theirs (data.storage; data.storageUnchecked is where it could not look). A "+
			"database started on old data keeps the password it was made with. It takes what "+
			"plan_install_app takes, and needs a key that may create apps in the env.",
		installInput,
		func(ctx context.Context, call *Call, in installArgs) (*apptemplatedto.PreflightAppFromTemplateResp, error) {
			ref, body, err := in.resolve(ctx, call)
			if err != nil {
				return nil, err
			}
			return preflight(ctx, call, ref, body)
		})
}

// preflight asks the create endpoint's preflight about a request, as the caller.
func preflight(ctx context.Context, call *Call, ref *envRef, body *apptemplatedto.CreateAppFromTemplateReq) (
	*apptemplatedto.PreflightAppFromTemplateResp, error) {
	var resp apptemplatedto.PreflightAppFromTemplateResp
	if err := call.Post(ctx, ref.path("/apps/from-template/preflight"), body, &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		resp.Data = &apptemplatedto.PreflightAppResult{}
	}
	return &resp, nil
}
