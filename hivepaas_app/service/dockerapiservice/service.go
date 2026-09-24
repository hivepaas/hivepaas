package dockerapiservice

import (
	"context"

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
}
