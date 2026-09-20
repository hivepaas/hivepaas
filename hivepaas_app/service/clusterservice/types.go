package clusterservice

import (
	"github.com/moby/moby/api/types/network"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type PersistingClusterData struct {
	UpsertingSettings []*entity.Setting
}

// PortRef is one address a service asks the cluster to answer at: the port a
// client connects to, and the protocol it speaks. A port is only taken for one
// protocol - 53/tcp and 53/udp are two different addresses.
type PortRef struct {
	Published uint32
	Protocol  network.IPProtocol
}
