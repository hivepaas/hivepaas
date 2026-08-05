package clusterservice

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/services/docker"
)

type Service interface {
	PersistClusterData(ctx context.Context, db database.IDB, data *PersistingClusterData) error

	// Docker services
	ServiceInspect(ctx context.Context, serviceID string, caching bool) (*swarm.Service, error)
	ServiceUpdate(ctx context.Context, serviceID string, version *swarm.Version, service *swarm.ServiceSpec,
		options ...docker.ServiceUpdateOption) (*client.ServiceUpdateResult, error)
	ServiceRemove(ctx context.Context, serviceID string, retryMax int, retryDelay time.Duration) error
	ServicesRemove(ctx context.Context, serviceIDs []string, retryMax int, retryDelay time.Duration) error

	// Docker nodes
	IsMultiNode(ctx context.Context) (bool, error)
	SyncNodes(ctx context.Context, db database.IDB) ([]swarm.Node, error)

	// Event handlers
	OnNodeEvent(ctx context.Context, event *events.Message) error
}
