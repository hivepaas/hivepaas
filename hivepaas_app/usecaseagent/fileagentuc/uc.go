// Package fileagentuc is the agent's side of files on volumes: it reads, writes,
// stats and removes a file inside a volume's directory on this node, and nothing
// outside it.
package fileagentuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

type UC struct {
	logger logging.Logger
	// hostPrefix is where the agent sees the node's filesystem.
	hostPrefix string
}

func New(logger logging.Logger) *UC {
	return &UC{logger: logger, hostPrefix: volumeservice.HostPathPrefix}
}

// NewOnHost is New with the node's filesystem seen at hostPrefix rather than at
// /host: an agent whose host is mounted elsewhere, or a test's directory.
func NewOnHost(logger logging.Logger, hostPrefix string) *UC {
	return &UC{logger: logger, hostPrefix: hostPrefix}
}
