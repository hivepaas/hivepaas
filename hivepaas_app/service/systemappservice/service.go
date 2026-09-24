// Package systemappservice runs the apps HivePaaS provisions for itself - the
// registry, the logging stack - in the hidden hivepaas project.
//
// They are ordinary apps. What makes one a system app is only who creates it and
// what owns its configuration: a system settings page, through the service that
// feature has. This package is the part those services share: where the app
// lives, how it is created, how a changed configuration reaches it, and how it
// goes away.
package systemappservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// LoadApp returns the system app with the key, or nil when it does not exist.
	// Settings, Project and ProjectEnv are loaded.
	//
	// The key is the identity rather than a stored id: an id left over from an app
	// somebody deleted must not stop the app from being created again.
	LoadApp(ctx context.Context, db database.IDB, key string) (*entity.App, error)

	// Provision creates the app. Its first deployment and certificate tasks come
	// back unscheduled: a task row can be picked up only once the transaction it
	// was written in has committed, and that transaction is the caller's.
	Provision(ctx context.Context, db database.IDB, req *ProvisionReq) (*ProvisionResp, error)

	// Redeploy changes the app's deployment settings and queues the deployment
	// that applies them. The task comes back unscheduled, for the same reason as
	// Provision's; it is nil when change reported nothing to do.
	Redeploy(ctx context.Context, db database.IDB, req *RedeployReq) (*entity.Task, error)

	// RecordImage writes the image the app's service now runs into its deployment
	// settings, for a caller that moved the service itself - the system updater,
	// which does it with a monitored rollback a deployment does not have. Without
	// it the next deployment of the app would put the old image back.
	RecordImage(ctx context.Context, db database.IDB, app *entity.App, image string) error

	// SyncSecrets makes the app's secrets exactly files: it creates the ones that
	// are missing, replaces the ones whose value or path changed and removes the
	// rest, in the swarm and in the app's settings. Every change is a service
	// update, so the app restarts once when anything changed.
	SyncSecrets(ctx context.Context, db database.IDB, app *entity.App, files []*SecretFile) error

	// Remove deletes the app and the apps created to serve it. removeStorage also
	// deletes its directories inside the volumes it mounted.
	Remove(ctx context.Context, db database.IDB, app *entity.App, removeStorage bool) error
}
