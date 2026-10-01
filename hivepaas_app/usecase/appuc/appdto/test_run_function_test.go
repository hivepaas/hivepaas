package appdto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

func testRunReq(request *TestRunRequestReq) *TestRunFunctionReq {
	return &TestRunFunctionReq{
		ProjectID:    "01J0000000000000000000PRJ1",
		ProjectEnvID: "01J0000000000000000000ENV1",
		AppID:        "01J0000000000000000000APP1",
		Code: &appsettingsdto.FunctionInlineCodeReq{Files: []*appsettingsdto.FunctionFileReq{
			{Path: "./index.js", Content: "export default () => ({})"},
		}},
		Request: request,
	}
}

// A test run takes the code as the editor has it and a request, which says
// only what it needs to: GET / by default.
func TestATestRunTakesTheCodeAndARequest(t *testing.T) {
	req := testRunReq(nil)

	assert.NoError(t, req.ModifyRequest())

	assert.Empty(t, req.Validate())
	assert.Equal(t, "index.js", req.Code.Files[0].Path)
	assert.Equal(t, "GET", req.Request.Method)
	assert.Equal(t, "/", req.Request.Path)
	run := req.ToRunRequest()
	assert.Equal(t, "GET", run.Method)
	assert.Equal(t, "index.js", req.Code.ToEntity().Files[0].Path)
}

func TestATestRunRequestIsNormalized(t *testing.T) {
	req := testRunReq(&TestRunRequestReq{Method: " post ", Path: "items/7", Body: `{"a":1}`,
		Headers: map[string][]string{"Content-Type": {"application/json"}}})

	assert.NoError(t, req.ModifyRequest())

	assert.Empty(t, req.Validate())
	run := req.ToRunRequest()
	assert.Equal(t, "POST", run.Method)
	assert.Equal(t, "/items/7", run.Path)
	assert.Equal(t, []byte(`{"a":1}`), run.Body)
	assert.Equal(t, []string{"application/json"}, run.Headers["content-type"], "header names in lower case")
}

func TestWhatATestRunCannotTakeIsRefused(t *testing.T) {
	for name, tc := range map[string]struct {
		req  *TestRunFunctionReq
		want string
	}{
		"no code": {func() *TestRunFunctionReq {
			req := testRunReq(nil)
			req.Code = nil
			return req
		}(), "code"},
		"a path outside the function": {func() *TestRunFunctionReq {
			req := testRunReq(nil)
			req.Code.Files[0].Path = "../index.js"
			return req
		}(), "code.files[0].path"},
		"a method HTTP does not have": {testRunReq(&TestRunRequestReq{Method: "BREW"}), "request.method"},
		"a body over 1 MB": {testRunReq(&TestRunRequestReq{Body: strings.Repeat("x", int(unit.MB)+1)}),
			"request.body"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, tc.req.ModifyRequest())
			assert.Contains(t, pathsOf(tc.req.Validate()), tc.want)
		})
	}
}
