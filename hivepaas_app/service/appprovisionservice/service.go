// Package appprovisionservice creates apps together with their configuration.
// An empty app is the smallest case; an app from a template is the full one.
package appprovisionservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// ProvisionApp creates an app, its swarm service and its settings inside the
	// caller's transaction.
	//
	// The swarm service cannot take part in that transaction. A failure after it
	// is created removes it before ProvisionApp returns; a failure the caller hits
	// afterwards is the caller's to clean up, through the returned app's ServiceID.
	//
	// TODO: spec import - import provisions each app of a bundle through this.
	ProvisionApp(ctx context.Context, db database.IDB, req *ProvisionAppReq) (*ProvisionAppResp, error)
}
