package dockerhelper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
	"github.com/hivepaas/hivepaas/services/docker"
)

// ErrHealthcheckMode is a healthcheck mode docker does not know.
var ErrHealthcheckMode = errors.New("the mode is not one of CMD, CMD-SHELL, NONE")

// CommandLine writes an argv as one command line, each word quoted as a shell
// needs it, so that CommandArgs gives the argv back: `sh -c 'a && b'`, where a
// plain join would make `sh -c a && b` of it.
func CommandLine(argv []string) string {
	return shellquote.Join(argv...)
}

// CommandArgs splits a command line by shell rules; nil for an empty one. A
// quote left open is an error.
func CommandArgs(line string) ([]string, error) {
	if strings.TrimSpace(line) == "" {
		return nil, nil
	}
	args, err := executil.CmdSplit(line)
	return args, hperrors.Wrap(err)
}

// ContainerCommandApply sets a container's entrypoint and its command, each
// from a command line. An empty one leaves the image's.
func ContainerCommandApply(contSpec *swarm.ContainerSpec, entrypoint, command string) error {
	entrypointArgs, err := CommandArgs(entrypoint)
	if err != nil {
		return err
	}
	commandArgs, err := CommandArgs(command)
	if err != nil {
		return err
	}
	contSpec.Command, contSpec.Args = entrypointArgs, commandArgs
	return nil
}

// HealthcheckTest is a healthcheck's test as docker runs it. CMD is an argv, so
// its command is split; CMD-SHELL is one string handed to the shell, so it is
// not. An empty mode inherits the image's test - nil, the timings still the
// healthcheck's own - unless it has a command: then it is CMD-SHELL, what
// someone writing a shell command expects. NONE turns the image's own off.
func HealthcheckTest(mode docker.HealthcheckMode, command string) ([]string, error) {
	switch mode {
	case docker.HealthcheckModeCmd:
		args, err := CommandArgs(command)
		if err != nil {
			return nil, err
		}
		return append([]string{string(mode)}, args...), nil
	case docker.HealthcheckModeInherit:
		if strings.TrimSpace(command) == "" {
			return nil, nil
		}
		return []string{string(docker.HealthcheckModeCmdShell), command}, nil
	case docker.HealthcheckModeCmdShell:
		return []string{string(docker.HealthcheckModeCmdShell), command}, nil
	case docker.HealthcheckModeNone:
		return []string{string(docker.HealthcheckModeNone)}, nil
	}
	return nil, fmt.Errorf("%w: %q", ErrHealthcheckMode, mode)
}

// HealthcheckCommand is a test's mode and its command line, HealthcheckTest the
// other way: CMD's argv quoted, CMD-SHELL's string as it is.
func HealthcheckCommand(test []string) (docker.HealthcheckMode, string) {
	if len(test) == 0 {
		return docker.HealthcheckModeInherit, ""
	}
	mode, rest := docker.HealthcheckMode(test[0]), test[1:]
	if mode == docker.HealthcheckModeCmd {
		return mode, CommandLine(rest)
	}
	// CMD-SHELL holds one string. One split into words before is shown
	// joined, as it was typed.
	return mode, strings.Join(rest, " ")
}
