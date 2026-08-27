package nodeagentuc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/nodeagentuc/nodeagentdto"
)

const (
	gracefulKillTimeout = 3 * time.Second
)

func (uc *UC) ExecuteCommand(
	ctx context.Context,
	req *nodeagentdto.ExecCommandReq,
) (*nodeagentdto.ExecCommandResp, error) {
	if req == nil || req.CommandExecOpts == nil || len(req.Command) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrBadRequest).WithExtraDetail("command cannot be empty")
	}

	uc.logger.Infof("ExecuteCommand started: %v, workingDir: %s", req.Command, req.WorkingDir)

	cmd := exec.Command(req.Command[0], req.Command[1:]...) //nolint:gosec
	if len(req.Env) > 0 {
		cmd.Env = append(os.Environ(), req.Env...)
	}
	cmd.Dir = req.WorkingDir

	cmd.Stdout = req.Stdout
	if req.Stderr != nil {
		cmd.Stderr = req.Stderr
	} else if req.Stdout != nil {
		cmd.Stderr = req.Stdout
	}

	// Create a new process group so that we can kill all child processes on interrupt
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		uc.logger.Errorf("Failed to start command %v: %v", req.Command, err)
		return nil, apperrors.Wrap(err)
	}

	doneChan := make(chan struct{})
	defer close(doneChan)

	// Watch for remote interrupt / context cancellation
	go func() {
		select {
		case <-ctx.Done():
			if cmd.Process != nil && cmd.Process.Pid > 0 {
				pgid := -cmd.Process.Pid
				uc.logger.Warnf("Context canceled, interrupting process group PGID: %d for command: %v",
					cmd.Process.Pid, req.Command)

				// First attempt graceful termination
				_ = syscall.Kill(pgid, syscall.SIGTERM)

				select {
				case <-time.After(gracefulKillTimeout):
					_ = syscall.Kill(pgid, syscall.SIGKILL)
				case <-doneChan:
					return
				}
			}
		case <-doneChan:
			return
		}
	}()

	waitErr := cmd.Wait()

	resp := &nodeagentdto.ExecCommandResp{}

	if cmd.ProcessState != nil {
		resp.ExitCode = int32(cmd.ProcessState.ExitCode()) //nolint:gosec
	}

	if waitErr != nil {
		if ctx.Err() != nil {
			uc.logger.Warnf("ExecuteCommand context canceled: %v", ctx.Err())
			return resp, apperrors.Wrap(ctx.Err())
		}
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			// Normal process exit with non-zero status
			return resp, nil
		}
		return resp, apperrors.Wrap(waitErr)
	}

	return resp, nil
}
