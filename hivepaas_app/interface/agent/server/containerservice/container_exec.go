package containerservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

type grpcExecStream struct {
	stream agentproto.ContainerService_ContainerExecServer
}

func (s *grpcExecStream) Context() context.Context {
	return s.stream.Context()
}

func (s *grpcExecStream) Recv() (*containeragentdto.ExecInput, error) {
	req, err := s.stream.Recv()
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	input := &containeragentdto.ExecInput{}

	if cfg := req.GetConfig(); cfg != nil {
		var consoleSize *containeragentdto.ConsoleSize
		if cfg.GetConsoleSize() != nil {
			consoleSize = &containeragentdto.ConsoleSize{
				Width:  cfg.GetConsoleSize().GetWidth(),
				Height: cfg.GetConsoleSize().GetHeight(),
			}
		}
		input.Config = &containeragentdto.ExecConfig{
			ContainerID: cfg.GetContainerId(),
			Cmd:         cfg.GetCmd(),
			Env:         cfg.GetEnv(),
			WorkingDir:  cfg.GetWorkingDir(),
			Tty:         cfg.GetTty(),
			ConsoleSize: consoleSize,
		}
	}

	if stdin := req.GetStdin(); stdin != nil {
		input.Stdin = stdin
	}

	if resize := req.GetResize(); resize != nil {
		input.Resize = &containeragentdto.ResizeOptions{
			Width:  resize.GetWidth(),
			Height: resize.GetHeight(),
		}
	}

	return input, nil
}

func (s *grpcExecStream) Send(out *containeragentdto.ExecOutput) error {
	resp := &agentproto.ContainerExecResp{}

	switch {
	case out.Stdout != nil:
		resp.Value = &agentproto.ContainerExecResp_Stdout{Stdout: out.Stdout}
	case out.Stderr != nil:
		resp.Value = &agentproto.ContainerExecResp_Stderr{Stderr: out.Stderr}
	case out.ExitCode != nil:
		resp.Value = &agentproto.ContainerExecResp_ExitCode{ExitCode: *out.ExitCode}
	}

	return s.stream.Send(resp) //nolint:wrapcheck
}

// ContainerExec starts a command execution in a container and pipes stdout/stderr/stdin/resize.
func ContainerExec(
	containerAgentUC *containeragentuc.UC,
	stream agentproto.ContainerService_ContainerExecServer,
) error {
	wrappedStream := &grpcExecStream{stream: stream}
	err := containerAgentUC.ExecuteCommand(wrappedStream)
	return apperrors.ToGRPCError(err) //nolint:wrapcheck
}
