package appdto

import (
	"net/http"
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

const (
	// testRunBodyMax is the largest body a test run sends: the request travels
	// to the build node with the code, in one message.
	testRunBodyMax = unit.MB
	testRunPathMax = 2048
)

var testRunMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodOptions}

// TestRunFunctionReq is a test run: the function's code as the editor has it,
// not yet saved, and the request to call it with.
type TestRunFunctionReq struct {
	ProjectID    string                                `json:"-"`
	ProjectEnvID string                                `json:"-"`
	AppID        string                                `json:"-"`
	Code         *appsettingsdto.FunctionInlineCodeReq `json:"code"`
	Request      *TestRunRequestReq                    `json:"request"`
}

// TestRunRequestReq is the request a test run calls the handler with. Its body
// is text.
type TestRunRequestReq struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Query   map[string][]string `json:"query"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

func NewTestRunFunctionReq() *TestRunFunctionReq {
	return &TestRunFunctionReq{}
}

// ModifyRequest implements interface basedto.ReqModifier: GET / unless the
// request says otherwise, header names in lower case as the runtime gives them.
func (req *TestRunFunctionReq) ModifyRequest() error {
	req.Code.Normalize()
	if req.Request == nil {
		req.Request = &TestRunRequestReq{}
	}
	r := req.Request
	r.Method = strings.ToUpper(strings.TrimSpace(r.Method))
	if r.Method == "" {
		r.Method = http.MethodGet
	}
	r.Path = strings.TrimSpace(r.Path)
	if !strings.HasPrefix(r.Path, "/") {
		r.Path = "/" + r.Path
	}
	if r.Headers != nil {
		headers := make(map[string][]string, len(r.Headers))
		for name, values := range r.Headers {
			name = strings.ToLower(strings.TrimSpace(name))
			headers[name] = append(headers[name], values...)
		}
		r.Headers = headers
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *TestRunFunctionReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, vld.Must(req.Code != nil).OnError(
		vld.SetField("code", nil),
		vld.SetCustomKey("ERR_VLD_VALUE_REQUIRED"),
	))
	if req.Code != nil {
		validators = append(validators, req.Code.Validate("code")...)
	}
	if r := req.Request; r != nil {
		validators = append(validators, vld.Must(slices.Contains(testRunMethods, r.Method)).OnError(
			vld.SetField("request.method", nil),
			vld.SetCustomKey("ERR_VLD_VALUE_NOT_IN_LIST"),
		))
		validators = append(validators, basedto.ValidateStr(&r.Path, true, 1, testRunPathMax, "request.path")...)
		validators = append(validators, basedto.ValidateStr(&r.Body, false, 0, int(testRunBodyMax),
			"request.body")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

// ToRunRequest is the request as a test run sends it.
func (req *TestRunFunctionReq) ToRunRequest() *functiontest.Request {
	r := req.Request
	return &functiontest.Request{
		Method:  r.Method,
		Path:    r.Path,
		Query:   r.Query,
		Headers: r.Headers,
		Body:    []byte(r.Body),
	}
}

type TestRunFunctionResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *TestRunFunctionDataResp `json:"data"`
}

// TestRunFunctionDataResp is what a test run brought back. Body is base64, as
// a response may not be text; BodyTruncated and LogsTruncated say what was cut
// at 1 MB. LockFiles are lock files the run made, for the editor to add.
type TestRunFunctionDataResp struct {
	Outcome        string                             `json:"outcome"`
	Status         int                                `json:"status,omitempty"`
	Headers        map[string][]string                `json:"headers,omitempty"`
	Body           []byte                             `json:"body,omitempty"`
	BodyTruncated  bool                               `json:"bodyTruncated,omitempty"`
	RequestID      string                             `json:"requestId,omitempty"`
	DurationMs     float64                            `json:"durationMs,omitempty"`
	Logs           string                             `json:"logs"`
	LogsTruncated  bool                               `json:"logsTruncated,omitempty"`
	Error          string                             `json:"error,omitempty"`
	ExitCode       int64                              `json:"exitCode"`
	LibrariesBuilt bool                               `json:"librariesBuilt"`
	LibrariesLog   string                             `json:"librariesLog,omitempty"`
	LockFiles      []*appsettingsdto.FunctionFileResp `json:"lockFiles,omitempty"`
}

// NULAllowed implements basedto.NULAllowed: a test run's request is passed to the function as it is.
func (req *TestRunFunctionReq) NULAllowed() {}
