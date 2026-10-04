package specdto

import (
	"regexp"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	composeMaxLen       = 1 << 20
	composeNameMaxLen   = 100
	composeEnvMaxLen    = 50
	composeDomainMaxLen = 253
	composeAppMaxLen    = 50
	composeColorMaxLen  = 20
	// composeDefaultEnv is the env a compose file's project is created with,
	// unless the request names another.
	composeDefaultEnv = "production"
)

// reComposeEnv is an env's name, as a project's create takes one.
var reComposeEnv = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// ValidateComposeReq asks what creating a project from a compose file would
// do, or adding its services to a project's env. Everything travels with each
// call, as an import's bundle does: nothing is kept on the server between
// validate and apply.
type ValidateComposeReq struct {
	// ProjectID is the project the services go into, from the route; empty
	// creates one.
	ProjectID string `json:"-"`

	Compose string `json:"compose"`
	// DotEnv is the .env beside the compose file.
	DotEnv string `json:"dotEnv"`
	// Files are the files the compose file reads, by their path in it. JSON
	// carries each as base64.
	Files map[string][]byte `json:"files"`
	// Variables are what the review gives of the file's variables, by name.
	Variables map[string]*ComposeVariableReq `json:"variables"`
	Project   ComposeProjectReq              `json:"project"`
	// Profiles are the profiles whose services are created, beside those with
	// none.
	Profiles []string `json:"profiles"`
	// Services are the review's choices of each service, by its name.
	Services map[string]*ComposeServiceReq `json:"services"`
	// Volume is the cluster volume the services' data goes to; none for the
	// project's own.
	Volume basedto.ObjectIDReq `json:"volume"`
	// Selection leaves services out, as an import's does: by their apps' paths.
	Selection specmodel.Selection `json:"selection"`
	// Deploy deploys the apps once they are created.
	Deploy bool `json:"deploy"`
}

type ComposeVariableReq struct {
	// Value is the variable's, over the .env's; absent keeps that one.
	Value *string `json:"value"`
	// Secret keeps it as an env secret; absent decides by its name.
	Secret *bool `json:"secret"`
}

type ComposeProjectReq struct {
	// Name is the project's; empty takes the file's own, `name:`. Not read
	// for an existing project.
	Name string `json:"name"`
	// Env is the env the services are created in, by its name.
	Env string `json:"env"`
	// NewEnv creates Env in an existing project, with EnvColor; otherwise the
	// project has it. A new project's env is always new.
	NewEnv   bool   `json:"newEnv"`
	EnvColor string `json:"envColor"`
}

type ComposeServiceReq struct {
	// Image is the image of a service the file only builds.
	Image string `json:"image"`
	// App is the app key chosen for the service; empty for its name's.
	App string `json:"app"`
	// UseExisting uses the existing env's app the service's name or key is,
	// rather than creating one.
	UseExisting bool              `json:"useExisting"`
	Ports       []*ComposePortReq `json:"ports"`
}

// ComposePortReq is the review's choice for a published port, found by
// Published, Target and Protocol.
type ComposePortReq struct {
	Published uint32                `json:"published"`
	Target    uint32                `json:"target"`
	Protocol  string                `json:"protocol"`
	As        composeservice.PortAs `json:"as"`
	Domain    string                `json:"domain"`
}

func NewValidateComposeReq() *ValidateComposeReq {
	return &ValidateComposeReq{}
}

// ModifyRequest defaults the env to production, as most people deploy a compose
// file to run it.
func (req *ValidateComposeReq) ModifyRequest() error {
	req.Project.Name = strings.TrimSpace(req.Project.Name)
	req.Project.Env = strings.TrimSpace(req.Project.Env)
	if req.Project.Env == "" {
		req.Project.Env = composeDefaultEnv
	}
	return nil
}

var _ basedto.ReqModifier = (*ValidateComposeReq)(nil)

func (req *ValidateComposeReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Compose, true, 1, composeMaxLen, "compose")...)
	validators = append(validators, basedto.ValidateStr(&req.Project.Name, false, 1, composeNameMaxLen,
		"project.name")...)
	validators = append(validators, basedto.ValidateStr(&req.Project.Env, true, 1, composeEnvMaxLen, "project.env")...)
	validators = append(validators, vld.Must(reComposeEnv.MatchString(req.Project.Env)).OnError(
		vld.SetField("project.env", nil),
		vld.SetCustomKey("ERR_VLD_PROJECT_ENV_NAME_INVALID"),
	))
	validators = append(validators, basedto.ValidateStr(&req.Project.EnvColor, false, 1, composeColorMaxLen,
		"project.envColor")...)
	validators = append(validators, basedto.ValidateObjectIDReq(&req.Volume, false, "volume")...)
	for name, svc := range req.Services {
		if svc == nil {
			continue
		}
		validators = append(validators, basedto.ValidateStr(&svc.App, false, 1, composeAppMaxLen,
			"services."+name+".app")...)
		for _, port := range svc.Ports {
			if port == nil {
				continue
			}
			field := "services." + name + ".ports"
			validators = append(validators, basedto.ValidateStrIn(&port.As, true, composeservice.AllPortAs,
				field+".as")...)
			validators = append(validators, basedto.ValidateStr(&port.Domain, false, 1, composeDomainMaxLen,
				field+".domain")...)
		}
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ApplyComposeReq creates what a validate of the same body planned. PlanHash
// is the plan the operator saw: a plan that changed since is refused.
type ApplyComposeReq struct {
	ValidateComposeReq
	PlanHash string `json:"planHash"`
	// AcceptIssues accepts every skipped, fixable and warning issue of the plan.
	AcceptIssues bool `json:"acceptIssues"`
}

func NewApplyComposeReq() *ApplyComposeReq {
	return &ApplyComposeReq{}
}

func (req *ApplyComposeReq) Validate() hperrors.ValidationErrors {
	errs := req.ValidateComposeReq.Validate()
	validators := basedto.ValidateStr(&req.PlanHash, true, 1, planHashMaxLen, "planHash")
	return append(errs, hperrors.NewValidationErrors(vld.Validate(validators...))...)
}

type ValidateComposeResp struct {
	Meta *basedto.Meta        `json:"meta"`
	Data *ValidateComposeData `json:"data"`
}

// ValidateComposeData is what the review shows: the project, each service as
// the app it becomes, the variables and files the file needs, and the import's
// plan - none while a required variable has no value, or a file an include
// reads is missing.
type ValidateComposeData struct {
	Project   *ComposeProjectResp            `json:"project"`
	Services  []*composeservice.ServiceView  `json:"services"`
	Variables []*composeservice.VariableView `json:"variables"`
	Needs     []*composeservice.FileNeed     `json:"needs"`
	Profiles  []string                       `json:"profiles"`
	Plan      *specmodel.ImportPlan          `json:"plan"`
}

type ComposeProjectResp struct {
	// ID is an existing project's; empty for one created.
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
	Env  string `json:"env"`
	// EnvKey is the env's key, as its paths in the plan name it.
	EnvKey string `json:"envKey"`
	// NewEnv says the env is created.
	NewEnv bool `json:"newEnv"`
	// FileName is the project's name in the file; empty for none.
	FileName string `json:"fileName"`
}

type ApplyComposeResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *ApplyComposeData `json:"data"`
}

type ApplyComposeData struct {
	// Project is the project created, or the one the services went into.
	Project *basedto.ObjectIDResp `json:"project"`
	// Plan is the plan applied, each selected node with its outcome.
	Plan *specmodel.ImportPlan `json:"plan"`
	// Deployments are the deployments queued.
	Deployments []*specservice.ImportDeployment `json:"deployments"`
}
