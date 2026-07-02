package containeragentuc

import (
	"context"
	"io"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

func (uc *UC) ExecuteCommand(
	stream containeragentdto.ExecStream,
) error {
	// 1. Receive the initial configuration message
	req, err := stream.Recv()
	if err != nil {
		return apperrors.New(err)
	}
	cfgMsg := req.Config
	if cfgMsg == nil {
		return apperrors.New(apperrors.ErrBadRequest).WithExtraDetail("first message must be ContainerCommandExecConfig")
	}

	uc.logger.Infof("ContainerCommandExec started in container: %s, cmd: %v",
		cfgMsg.ContainerID, cfgMsg.Cmd)

	// 2. Configure Docker Exec Options
	var actualExecOpts *client.ExecCreateOptions
	execOptions := func(opts *client.ExecCreateOptions) {
		actualExecOpts = opts
		opts.AttachStdin = true
		opts.AttachStdout = true
		opts.AttachStderr = true
		opts.Cmd = cfgMsg.Cmd
		opts.Env = cfgMsg.Env
		opts.WorkingDir = cfgMsg.WorkingDir
		opts.TTY = cfgMsg.Tty
		if cfgMsg.ConsoleSize != nil {
			opts.ConsoleSize.Width = uint(cfgMsg.ConsoleSize.Width)
			opts.ConsoleSize.Height = uint(cfgMsg.ConsoleSize.Height)
		}
	}

	// 3. Trigger Docker Exec Creation and Attachment
	createResp, attachResp, _, err := uc.dockerManager.ContainerExec(stream.Context(),
		cfgMsg.ContainerID, execOptions)
	if err != nil {
		uc.logger.Errorf("Failed to initialize container exec: %v", err)
		return apperrors.New(apperrors.ErrInternal).WithCause(err).WithExtraDetail("Docker exec failed")
	}
	defer attachResp.Close()

	// 4. Start Goroutine B: gRPC client stdin/resize -> Docker exec input
	go func() {
		for {
			inReq, recvErr := stream.Recv()
			if recvErr != nil {
				return // Stream closed or EOF
			}
			if len(inReq.Stdin) > 0 && attachResp.Conn != nil {
				_, _ = attachResp.Conn.Write(inReq.Stdin)
			}
			if inReq.Resize != nil {
				_, _ = uc.dockerManager.ContainerExecResize(context.Background(), createResp.ID,
					uint(inReq.Resize.Width), uint(inReq.Resize.Height))
			}
		}
	}()

	// 5. Pipe Docker exec output -> gRPC stream (Tty vs non-Tty multiplexing)
	var copyErr error
	if actualExecOpts.TTY { // NOTE: do not use cfgMsg.GetTty()
		// Tty mode combines both stdout and stderr
		_, copyErr = io.Copy(&streamWriter{stream: stream, isStdout: true}, attachResp.Reader)
	} else {
		// Non-Tty mode uses Docker stdcopy header protocol to multiplex streams
		_, copyErr = stdcopy.StdCopy(
			&streamWriter{stream: stream, isStdout: true},
			&streamWriter{stream: stream, isStdout: false},
			attachResp.Reader,
		)
	}
	if copyErr != nil {
		uc.logger.Warnf("Error copying terminal buffer: %v", copyErr)
	}

	// 6. Inspect the completed process to retrieve its ExitCode
	var exitCode int32 = 0
	inspectResp, inspectErr := uc.dockerManager.ContainerExecInspect(stream.Context(), createResp.ID)
	if inspectErr == nil {
		exitCode = int32(inspectResp.ExitCode) //nolint:gosec
	}

	// 7. Send the exit code as the final message
	_ = stream.Send(&containeragentdto.ExecOutput{
		ExitCode: &exitCode,
	})

	uc.logger.Infof("ContainerCommandExec finished in container %s with exit code: %d",
		cfgMsg.ContainerID, exitCode)

	return nil
}

type streamWriter struct {
	stream   containeragentdto.ExecStream
	isStdout bool
}

func (w *streamWriter) Write(p []byte) (n int, err error) {
	var resp containeragentdto.ExecOutput
	// Copy buffer bytes to keep data thread-safe
	buf := make([]byte, len(p))
	copy(buf, p)

	if w.isStdout {
		resp.Stdout = buf
	} else {
		resp.Stderr = buf
	}

	if sendErr := w.stream.Send(&resp); sendErr != nil {
		return 0, apperrors.New(sendErr)
	}
	return len(p), nil
}
