package schedjobdto

import (
	"net/http"
	"regexp"
	"slices"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	maxFunctionInvokePathLen = 2048
	maxFunctionInvokeHeaders = 50
	// maxFunctionInvokeBody is a test run's: the request travels the same way.
	maxFunctionInvokeBody = int(unit.MB)
)

var (
	functionInvokeMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions}
	// functionInvokePathRegex is a path with its query, without a space or a
	// control character.
	functionInvokePathRegex = regexp.MustCompile(`^/[^\s\x00-\x1f\x7f]*$`)
	// functionInvokeHeaderNameRegex is an HTTP token, in lower case.
	functionInvokeHeaderNameRegex = regexp.MustCompile("^[a-z0-9!#$%&'*+.^_`|~-]+$")
)

// SchedJobFunctionInvokeReq is the request a function-invoke job calls its
// function with: the query in the path, the body as text.
type SchedJobFunctionInvokeReq struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

func (req *SchedJobFunctionInvokeReq) ToEntity() *entity.SchedJobFunctionInvoke {
	if req == nil {
		return nil
	}
	res := &entity.SchedJobFunctionInvoke{Method: req.Method, Path: req.Path, Headers: req.Headers, Body: req.Body}
	if len(res.Headers) == 0 {
		res.Headers = nil
	}
	return res
}

// modifyRequest spells the request one way: the method in upper case, GET when
// left out; the path from the root, "/" when left out; header names in lower
// case.
func (req *SchedJobFunctionInvokeReq) modifyRequest() {
	if req == nil {
		return
	}
	req.Method = strings.ToUpper(strings.TrimSpace(req.Method))
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	req.Path = strings.TrimSpace(req.Path)
	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/" + req.Path
	}
	headers := make(map[string][]string, len(req.Headers))
	for name, values := range req.Headers {
		name = strings.ToLower(strings.TrimSpace(name))
		headers[name] = append(headers[name], values...)
	}
	req.Headers = headers
}

func (req *SchedJobFunctionInvokeReq) validate(field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateCond(req != nil, field)...)
	if req == nil {
		return res
	}
	field += "."
	res = append(res, basedto.ValidateCond(slices.Contains(functionInvokeMethods, req.Method), field+"method")...)
	res = append(res, basedto.ValidateCond(len(req.Path) <= maxFunctionInvokePathLen &&
		functionInvokePathRegex.MatchString(req.Path), field+"path")...)
	res = append(res, basedto.ValidateCond(validFunctionInvokeHeaders(req.Headers), field+"headers")...)
	sendsBody := req.Method != http.MethodGet && req.Method != http.MethodHead
	res = append(res, basedto.ValidateCond(len(req.Body) <= maxFunctionInvokeBody && (sendsBody || req.Body == ""),
		field+"body")...)
	return res
}

func validFunctionInvokeHeaders(headers map[string][]string) bool {
	if len(headers) > maxFunctionInvokeHeaders {
		return false
	}
	for name, values := range headers {
		if !functionInvokeHeaderNameRegex.MatchString(name) {
			return false
		}
		for _, value := range values {
			if strings.ContainsAny(value, "\r\n\x00") {
				return false
			}
		}
	}
	return true
}

// validateFunctionInvokeFields is what a job's type says about its request: a
// function-invoke has one, runs no command of its own and names its app; any
// other type has none.
func (req *SchedJobBaseReq) validateFunctionInvokeFields(field string) (res []vld.Validator) {
	if req.JobType != base.SchedJobTypeFunctionInvoke {
		return basedto.ValidateCond(req.FunctionInvoke == nil, field+"functionInvoke")
	}
	res = append(res, req.FunctionInvoke.validate(field+"functionInvoke")...)
	res = append(res, basedto.ValidateCond(req.Command == nil, field+"command")...)
	res = append(res, basedto.ValidateCond(req.CommandOutput == nil, field+"commandOutput")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.App, true, field+"app")...)
	return res
}

type SchedJobFunctionInvokeResp struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
}
