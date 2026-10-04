package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

type composeInput struct {
	Compose string `json:"compose" jsonschema:"the docker-compose.yml, as text"`

	DotEnv string `json:"dotEnv,omitempty" jsonschema:"the .env beside it, as text"`

	Files map[string]string `json:"files,omitempty" jsonschema:"what the file reads, as text, by its path in it"`

	Variables map[string]composeVariableInput `json:"variables,omitempty" jsonschema:"variables, over the .env"`

	Project string `json:"project,omitempty" jsonschema:"the project's name; empty takes the file's"`

	Env string `json:"env,omitempty" jsonschema:"the env the apps are in; production when empty"`

	Profiles []string `json:"profiles,omitempty" jsonschema:"profiles whose services are created"`

	Ports map[string][]composePortInput `json:"ports,omitempty" jsonschema:"each service's ports' choices"`

	Images map[string]string `json:"images,omitempty" jsonschema:"images of services the file only builds"`

	Deploy *bool `json:"deploy,omitempty" jsonschema:"deploy the apps once created; true when not given"`
}

type composeVariableInput struct {
	Value  *string `json:"value,omitempty" jsonschema:"the variable's value"`
	Secret *bool   `json:"secret,omitempty" jsonschema:"keep it as an env secret; by its name when not given"`
}

type composePortInput struct {
	Published uint32 `json:"published" jsonschema:"the port compose publishes on the host"`
	Target    uint32 `json:"target" jsonschema:"the container's port"`
	Protocol  string `json:"protocol,omitempty" jsonschema:"tcp or udp; tcp when empty"`
	As        string `json:"as" jsonschema:"domain, node or none"`
	Domain    string `json:"domain,omitempty" jsonschema:"the domain, for as domain"`
}

type composePlan struct {
	// Review is what validate answers: each service as the app it becomes, the
	// variables and files the file needs, and the import's plan.
	Review *specdto.ValidateComposeData `json:"review"`
	Next   string                       `json:"next,omitempty"`
}

func planCreateProjectFromComposeTool() Tool {
	return planTool("plan_create_project_from_compose", "Plan a project from a compose file",
		"Plans POST /projects/from-compose/apply: a new project from a docker-compose.yml, one env, an app per "+
			"service reached by the same name. The plan is what POST /projects/from-compose/validate answers: "+
			"each service as the app it becomes (its image, what each published port becomes, its volumes, "+
			"what is left out), the variables (a required one with no value stops everything until it has "+
			"one), the files the compose file reads that were not given, and the import's plan with its "+
			"issues. Give a file's text in files, a variable's value in variables, a port's choice in ports. "+
			"A plan with a blocked issue cannot be applied. Applied, the project is created and its apps' "+
			"first deployments queued. Nothing is created until apply_plan.",
		NeedWrite, &applier{follow: "list_projects finds the new project; get_app_deployment with each " +
			"deployment id shows it pull and start."},
		func(ctx context.Context, call *Call, in composeInput) (composePlan, *storedPlan, error) {
			if strings.TrimSpace(in.Compose) == "" {
				return composePlan{}, nil, &InputError{Message: "compose is required: the compose file's text"}
			}
			req := in.request()
			var resp specdto.ValidateComposeResp
			if err := call.Post(ctx, "/projects/from-compose/validate", &req.ValidateComposeReq, &resp); err != nil {
				return composePlan{}, nil, err
			}
			out := composePlan{Review: resp.Data}
			plan := resp.Data.Plan
			if plan == nil {
				out.Next = "Give the required variables a value (review.variables, given false), and plan again."
				return out, nil, nil
			}
			if plan.Summary[string(specmodel.SeverityBlocked)] > 0 {
				out.Next = "An issue blocks it (review.plan): change the file or the choices, and plan again."
				return out, nil, nil
			}
			req.PlanHash = plan.PlanHash
			for _, severity := range []specmodel.Severity{specmodel.SeveritySkipped, specmodel.SeverityFixable,
				specmodel.SeverityWarning} {
				req.AcceptIssues = req.AcceptIssues || plan.Summary[string(severity)] > 0
			}
			raw, err := json.Marshal(req)
			if err != nil {
				return composePlan{}, nil, fmt.Errorf("mcp: encoding a plan: %w", err)
			}
			if req.AcceptIssues {
				out.Next = "Applying it accepts the plan's issues: tell the person each, as review.plan says."
			}
			return out, &storedPlan{Method: http.MethodPost, Path: "/projects/from-compose/apply", Body: raw,
				Summary: fmt.Sprintf("create the project %s from a compose file, %d service(s)",
					resp.Data.Project.Name, len(resp.Data.Services))}, nil
		})
}

// request is the endpoints' body of the input: the files' text as the bytes
// the API takes.
func (in composeInput) request() *specdto.ApplyComposeReq {
	req := specdto.NewApplyComposeReq()
	req.Compose, req.DotEnv, req.Profiles = in.Compose, in.DotEnv, in.Profiles
	req.Project = specdto.ComposeProjectReq{Name: strings.TrimSpace(in.Project), Env: strings.TrimSpace(in.Env)}
	req.Deploy = in.Deploy == nil || *in.Deploy
	req.Files = map[string][]byte{}
	for path, text := range in.Files {
		req.Files[path] = []byte(text)
	}
	req.Variables = map[string]*specdto.ComposeVariableReq{}
	for name, v := range in.Variables {
		req.Variables[name] = &specdto.ComposeVariableReq{Value: v.Value, Secret: v.Secret}
	}
	req.Services = map[string]*specdto.ComposeServiceReq{}
	service := func(name string) *specdto.ComposeServiceReq {
		if req.Services[name] == nil {
			req.Services[name] = &specdto.ComposeServiceReq{}
		}
		return req.Services[name]
	}
	for name, image := range in.Images {
		service(name).Image = strings.TrimSpace(image)
	}
	for name, ports := range in.Ports {
		for _, port := range ports {
			protocol := port.Protocol
			if protocol == "" {
				protocol = "tcp"
			}
			service(name).Ports = append(service(name).Ports, &specdto.ComposePortReq{
				Published: port.Published, Target: port.Target, Protocol: protocol,
				As: composeservice.PortAs(port.As), Domain: port.Domain,
			})
		}
	}
	return req
}
