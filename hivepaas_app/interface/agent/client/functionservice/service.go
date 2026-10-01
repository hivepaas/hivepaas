// Package functionservice calls an agent's function service: a test run on the
// agent's node.
package functionservice

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	clientbuild "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/imagebuildservice"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functiontest"
)

// maxAnswerSize is how large a test run's answer may be: a body and logs of a
// megabyte each, lock files, the install's log - over gRPC's default of 4 MB.
const maxAnswerSize = 16 << 20

type FunctionServiceClient interface {
	// TestRun runs a function's test run on the agent's node.
	TestRun(ctx context.Context, req *functiontest.RunReq) (*functiontest.RunResp, error)
	Close() error
}

type grpcFunctionServiceClient struct {
	protoClient agentproto.FunctionServiceClient
	conn        *grpc.ClientConn
}

func NewFunctionServiceClient(agentAddr string) (FunctionServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcFunctionServiceClient{conn: conn, protoClient: agentproto.NewFunctionServiceClient(conn)}, nil
}

func (c *grpcFunctionServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcFunctionServiceClient) TestRun(
	ctx context.Context,
	req *functiontest.RunReq,
) (*functiontest.RunResp, error) {
	// The agent cannot open a stored secret: what the install reads from
	// settings goes with the request, resolved here.
	if req.Inputs == nil {
		return nil, hperrors.NewMissing("Function test run inputs")
	}
	resp, err := c.protoClient.FunctionTestRun(client.CreateAuthCtx(ctx), runReqToProto(req),
		grpc.MaxCallRecvMsgSize(maxAnswerSize))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return runRespFromProto(resp), nil
}

func runReqToProto(req *functiontest.RunReq) *agentproto.FunctionTestRunReq {
	src := req.Source
	files := make([]*agentproto.FunctionFile, 0, len(req.Files))
	for _, f := range req.Files {
		files = append(files, &agentproto.FunctionFile{Path: f.Path, Content: f.Content})
	}
	out := &agentproto.FunctionTestRunReq{
		Source: &agentproto.FunctionSource{
			Runtime:           string(src.Runtime),
			Contract:          string(src.Contract),
			EntrypointFile:    src.Entrypoint.File,
			EntrypointHandler: src.Entrypoint.Handler,
			SystemPackages:    src.SystemPackages,
			TimeoutMs:         time.Duration(src.Timeout).Milliseconds(),
			MaxConcurrency:    int64(src.MaxConcurrency),
			MaxBodySize:       src.MaxBodySize.Bytes(),
		},
		Files:         files,
		Images:        req.Images,
		Inputs:        clientbuild.InputsToProto(req.Inputs),
		BuildSettings: clientbuild.BuildSettingsToProto(req.BuildSettings),
		Env:           req.Env,
		Network:       req.Network,
		NanoCpus:      req.NanoCPUs,
		MemoryBytes:   req.MemoryBytes,
	}
	if r := req.Request; r != nil {
		out.Request = &agentproto.FunctionRequest{
			Method: r.Method, Path: r.Path, Query: valuesToProto(r.Query), Headers: valuesToProto(r.Headers),
			Body: r.Body,
		}
	}
	return out
}

func runRespFromProto(resp *agentproto.FunctionTestRunResp) *functiontest.RunResp {
	var lockFiles []*entity.FunctionFile
	for _, f := range resp.GetLockFiles() {
		lockFiles = append(lockFiles, &entity.FunctionFile{Path: f.GetPath(), Content: f.GetContent()})
	}
	return &functiontest.RunResp{
		Outcome:        functiontest.Outcome(resp.GetOutcome()),
		Status:         int(resp.GetStatus()),
		Headers:        valuesFromProto(resp.GetHeaders()),
		Body:           resp.GetBody(),
		BodyTruncated:  resp.GetBodyTruncated(),
		RequestID:      resp.GetRequestId(),
		DurationMs:     resp.GetDurationMs(),
		Logs:           resp.GetLogs(),
		LogsTruncated:  resp.GetLogsTruncated(),
		Error:          resp.GetError(),
		ExitCode:       resp.GetExitCode(),
		LibrariesBuilt: resp.GetLibrariesBuilt(),
		LibrariesLog:   resp.GetLibrariesLog(),
		LockFiles:      lockFiles,
	}
}

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
