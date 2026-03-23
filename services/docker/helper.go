package docker

import (
	"strings"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
)

func ContainerCommandBuild(cmd []string, args []string) string {
	return strings.Join(gofn.Concat(cmd, args), " ")
}

func ContainerCommandApply(contSpec *swarm.ContainerSpec, cmd string) {
	if cmd == "" {
		contSpec.Command = nil
	} else {
		contSpec.Command = gofn.Must(shellutil.CmdSplit(cmd))
	}
}
