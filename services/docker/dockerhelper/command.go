package dockerhelper

import (
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/executil"
)

func ContainerCommandBuild(cmd []string, args []string) string {
	return strings.Join(gofn.Concat(cmd, args), " ")
}

func ContainerCommandApply(contSpec *swarm.ContainerSpec, cmd string) {
	contSpec.Command = nil
	if cmd == "" {
		contSpec.Args = nil
	} else {
		contSpec.Args = gofn.Must(executil.CmdSplit(cmd))
	}
}
