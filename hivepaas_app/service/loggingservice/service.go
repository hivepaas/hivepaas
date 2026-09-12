package loggingservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Status is what the logging stack is currently doing.
type Status struct {
	Enabled          bool
	BackendServiceID string
	CollectorService string
	BackendReady     bool
	// ExcludedApps are apps whose own log driver keeps them out of collection.
	ExcludedApps []string
}

type Service interface {
	// Apply makes the cluster match the stored configuration, deploying or
	// removing as needed.
	Apply(ctx context.Context, db database.IDB) error

	// TearDown removes the collector and the backend, keeping the data volume.
	TearDown(ctx context.Context) error

	Status(ctx context.Context, db database.IDB) (*Status, error)
}
