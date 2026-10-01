// Package functionservice serves the agent's function calls: a test run on its
// node.
package functionservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	serverbuild "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/functionagentuc"
)

// FunctionTestRun runs the test run the call carries, on this node.
func FunctionTestRun(
	ctx context.Context,
	uc *functionagentuc.UC,
	req *agentproto.FunctionTestRunReq,
) (*agentproto.FunctionTestRunResp, error) {
	resp, err := uc.TestRun(ctx, runReqFromProto(req))
	if err != nil {
		return nil, hperrors.ToGRPCError(err) //nolint:wrapcheck
	}
	return runRespToProto(resp), nil
}

func runReqFromProto(req *agentproto.FunctionTestRunReq) *functiontest.RunReq {
	src := req.GetSource()
	files := make([]*entity.FunctionFile, 0, len(req.GetFiles()))
	for _, f := range req.GetFiles() {
		files = append(files, &entity.FunctionFile{Path: f.GetPath(), Content: f.GetContent()})
	}
	request := req.GetRequest()
	return &functiontest.RunReq{
		Source: &entity.DeploymentFunctionSource{
			Runtime:        base.FunctionRuntime(src.GetRuntime()),
			Contract:       base.FunctionContract(src.GetContract()),
			Entrypoint:     entity.FunctionEntrypoint{File: src.GetEntrypointFile(), Handler: src.GetEntrypointHandler()},
			SystemPackages: src.GetSystemPackages(),
			Timeout:        timeutil.Duration(time.Duration(src.GetTimeoutMs()) * time.Millisecond),
			MaxConcurrency: int(src.GetMaxConcurrency()),
			MaxBodySize:    unit.DataSize(src.GetMaxBodySize()),
		},
		Files: files,
		Request: &functiontest.Request{
			Method:  request.GetMethod(),
			Path:    request.GetPath(),
			Query:   valuesFromProto(request.GetQuery()),
			Headers: valuesFromProto(request.GetHeaders()),
			Body:    request.GetBody(),
		},
		Images:        req.GetImages(),
		Inputs:        serverbuild.InputsFromProto(req.GetInputs()),
		BuildSettings: serverbuild.BuildSettingsFromProto(req.GetBuildSettings()),
		Env:           req.GetEnv(),
		Network:       req.GetNetwork(),
		NanoCPUs:      req.GetNanoCpus(),
		MemoryBytes:   req.GetMemoryBytes(),
	}
}

func runRespToProto(resp *functiontest.RunResp) *agentproto.FunctionTestRunResp {
	lockFiles := make([]*agentproto.FunctionFile, 0, len(resp.LockFiles))
	for _, f := range resp.LockFiles {
		lockFiles = append(lockFiles, &agentproto.FunctionFile{Path: f.Path, Content: f.Content})
	}
	return &agentproto.FunctionTestRunResp{
		Outcome:        string(resp.Outcome),
		Status:         int32(resp.Status), //nolint:gosec // an HTTP status
		Headers:        valuesToProto(resp.Headers),
		Body:           resp.Body,
		BodyTruncated:  resp.BodyTruncated,
		RequestId:      resp.RequestID,
		DurationMs:     resp.DurationMs,
		Logs:           resp.Logs,
		LogsTruncated:  resp.LogsTruncated,
		Error:          resp.Error,
		ExitCode:       resp.ExitCode,
		LibrariesBuilt: resp.LibrariesBuilt,
		LibrariesLog:   resp.LibrariesLog,
		LockFiles:      lockFiles,
	}
}

// valuesToProto is a query's or headers' values by name, as the agent's calls
// carry them; nil for none.
func valuesToProto(values map[string][]string) map[string]*agentproto.FunctionValues {
	if values == nil {
		return nil
	}
	out := make(map[string]*agentproto.FunctionValues, len(values))
	for name, v := range values {
		out[name] = &agentproto.FunctionValues{Values: v}
	}
	return out
}

func valuesFromProto(values map[string]*agentproto.FunctionValues) map[string][]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string][]string, len(values))
	for name, v := range values {
		out[name] = v.GetValues()
	}
	return out
}
