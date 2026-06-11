package executil

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
)

type CmdRes struct {
	Exit      int
	Error     error
	Code      string
	Output    []string
	RawOutput []string
}

func (res *CmdRes) Succeed() bool {
	return res.Exit == 0 && res.Error == nil
}

func (res *CmdRes) ErrorStr() string {
	if res.Error == nil {
		return ""
	}
	return res.Error.Error()
}

func Execute(cmd *exec.Cmd, cmdRes *CmdRes) {
	res, err := cmd.CombinedOutput()
	if err != nil {
		cmdRes.Error = errors.Join(cmdRes.Error, err)
	}
	if cmd.ProcessState != nil {
		cmdRes.Exit = cmd.ProcessState.ExitCode()
	}
	cmdRes.RawOutput = append(cmdRes.RawOutput, reflectutil.UnsafeBytesToStr(res))
	cmdRes.Output = append(cmdRes.Output, strings.Split(reflectutil.UnsafeBytesToStr(res), "\n")...)
}
