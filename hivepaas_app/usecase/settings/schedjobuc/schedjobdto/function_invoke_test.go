package schedjobdto

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const functionApp = "01JAB9XED0GTXBSQDFVYAJ8WB4"

func functionInvokeReq(invoke *SchedJobFunctionInvokeReq) *CreateSchedJobReq {
	req := NewCreateSchedJobReq()
	req.SchedJobBaseReq = &SchedJobBaseReq{
		Name:           "nightly report",
		JobType:        base.SchedJobTypeFunctionInvoke,
		App:            basedto.ObjectIDReq{ID: functionApp},
		FunctionInvoke: invoke,
	}
	return req
}

func TestAFunctionCallIsItsRequest(t *testing.T) {
	req := functionInvokeReq(&SchedJobFunctionInvokeReq{
		Method:  " post ",
		Path:    "reports/daily?format=csv",
		Headers: map[string][]string{"Content-Type": {"application/json"}, "X-Trace": {"a", "b"}},
		Body:    `{"day":"today"}`,
	})

	assert.Equal(t, "", invalidFields(t, req))
	job := req.ToEntity()
	if assert.NotNil(t, job.FunctionInvoke) {
		assert.Equal(t, "POST", job.FunctionInvoke.Method)
		assert.Equal(t, "/reports/daily?format=csv", job.FunctionInvoke.Path)
		assert.Equal(t, map[string][]string{"content-type": {"application/json"}, "x-trace": {"a", "b"}},
			job.FunctionInvoke.Headers, "header names are lower case, as invoke reads them")
		assert.Equal(t, `{"day":"today"}`, job.FunctionInvoke.Body)
	}
	assert.Nil(t, job.Command, "a call runs the runtime's invoke, not a command of the job's")
}

func TestAFunctionCallDefaultsToAGetOfTheRoot(t *testing.T) {
	req := functionInvokeReq(&SchedJobFunctionInvokeReq{})

	assert.Equal(t, "", invalidFields(t, req))
	job := req.ToEntity()
	assert.Equal(t, "GET", job.FunctionInvoke.Method)
	assert.Equal(t, "/", job.FunctionInvoke.Path)
	assert.Nil(t, job.FunctionInvoke.Headers)
}

func TestAFunctionCallRefusesWhatInvokeCannotSend(t *testing.T) {
	cases := map[string]*SchedJobFunctionInvokeReq{
		"functionInvoke.method":  {Method: "FETCH"},
		"functionInvoke.path":    {Path: "/a b"},
		"functionInvoke.path ":   {Path: "/" + strings.Repeat("a", 2048)},
		"functionInvoke.headers": {Headers: map[string][]string{"bad name": {"x"}}},
		"functionInvoke.body":    {Method: "POST", Body: strings.Repeat("a", 1024*1024+1)},
		"functionInvoke.body ":   {Method: "GET", Body: "a GET sends no body"},
	}
	for field, invoke := range cases {
		assert.Contains(t, invalidFields(t, functionInvokeReq(invoke)), strings.TrimSpace(field), field)
	}

	many := map[string][]string{}
	for i := range 51 {
		many[fmt.Sprintf("x-h%d", i)] = []string{"v"}
	}
	assert.Contains(t, invalidFields(t, functionInvokeReq(&SchedJobFunctionInvokeReq{Headers: many})),
		"functionInvoke.headers")
}

func TestAFunctionCallIsOnlyAFunctionInvokesAndNeedsItsApp(t *testing.T) {
	req := functionInvokeReq(nil)
	assert.Contains(t, invalidFields(t, req), "functionInvoke")

	req = functionInvokeReq(&SchedJobFunctionInvokeReq{})
	req.App = basedto.ObjectIDReq{}
	assert.Contains(t, invalidFields(t, req), "app")

	req = functionInvokeReq(&SchedJobFunctionInvokeReq{})
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "echo hi"}
	assert.Contains(t, invalidFields(t, req), "command")

	req = functionInvokeReq(&SchedJobFunctionInvokeReq{})
	req.JobType = base.SchedJobTypeContainerCommand
	req.Command = &commandtemplatedto.CommandTemplateBaseReq{Command: "echo hi"}
	assert.Contains(t, invalidFields(t, req), "functionInvoke")
}
