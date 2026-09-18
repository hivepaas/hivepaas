// Package appprovisionservice creates apps together with their configuration.
// An empty app is the smallest case; an app from a template is the full one.
//
// It is the counterpart of appcloneservice: both create an app that did not
// exist, one from a template or from nothing, the other from an app that does.
// What the two have in common - applying to docker what an app's settings say -
// lives here, in ApplyAppConfiguration, and cloning calls it on the copy.
package appprovisionservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// ProvisionApp creates an app, its swarm service and its settings inside the
	// caller's transaction, applies the configuration those settings describe,
	// and creates the first deployment when the request asks for one.
	//
	// The swarm service cannot take part in that transaction. A failure after it
	// is created removes it before ProvisionApp returns, along with the docker
	// secrets and configs created for it; a failure the caller hits afterwards is
	// the caller's to clean up, through the returned response.
	//
	// TODO: spec import - import provisions each app of a bundle through this.
	ProvisionApp(ctx context.Context, db database.IDB, req *ProvisionAppReq) (*ProvisionAppResp, error)

	// ProvisionApps provisions several apps in the order given, as one creation:
	// a template that brings a database along creates both this way, and each app
	// can name the others because their ids are chosen before the first exists.
	//
	// It returns what it created even when it fails part way, so that the caller
	// whose transaction is about to roll back can undo it through Cleanup.
	ProvisionApps(ctx context.Context, db database.IDB, req *ProvisionAppsReq) (*ProvisionAppsResp, error)

	// ApplyAppConfiguration applies to docker what an app's settings say: its
	// environment, config files, secrets, routing and scheduled jobs. Each step
	// does nothing when the app has no setting of its kind.
	//
	// It runs on an app whose settings are already persisted, and writes the
	// settings again where docker hands back something they have to keep - the
	// id a secret was created with, without which nothing could find it again.
	ApplyAppConfiguration(ctx context.Context, db database.IDB, req *ApplyAppConfigurationReq) (
		*ApplyAppConfigurationResp, error)
}
