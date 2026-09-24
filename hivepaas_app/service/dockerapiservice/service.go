package dockerapiservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
)

type Service interface {
	// Policies are what every app with Docker API access may do, read from the
	// database. Every node's agent serves exactly these.
	Policies(ctx context.Context, db database.IDB) ([]*dockerproxy.Policy, error)
	// SyncAgents asks every node's agent to serve exactly the apps that have
	// access now, rather than at its next tick.
	SyncAgents(ctx context.Context) error
	// RemoveAppObjects asks every node's agent to stop serving an app and to
	// remove what its children left: containers, networks and volumes.
	RemoveAppObjects(ctx context.Context, appID string) error
	// AccessOf is an app's Docker API setting, or nil when it has none.
	AccessOf(ctx context.Context, db database.IDB, appID string) (*entity.AppDockerAPISettings, error)
	// EnsureNetwork returns the id of the app's own network, creating it first
	// when it does not exist.
	EnsureNetwork(ctx context.Context, appID string) (string, error)
	// AppNetworkID is the id of the app's own network, "" when it has none.
	AppNetworkID(ctx context.Context, appID string) (string, error)
	// ApplyToService gives an app's service spec what its access needs - its
	// socket and its network - or takes them away when it has none.
	ApplyToService(ctx context.Context, db database.IDB, appID string, spec *swarm.ServiceSpec) error
	// DetachFromService takes an app's socket and network off a spec: a clone's,
	// which starts as a copy of that app's.
	DetachFromService(ctx context.Context, appID string, spec *swarm.ServiceSpec) error
	// RemoveApp removes what an app's access made: its children and socket
	// volumes on every node, and its network. Its setting goes with the app's.
	RemoveApp(ctx context.Context, appID string) error
}
