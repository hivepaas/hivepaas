// Package managedservice is the engines that look after managed templates' apps.
package managedservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
)

type Service interface {
	// Engine is the engine of that name, and false when this HivePaaS has none.
	Engine(name templatemodel.ManagedEngine) (managedmodel.Engine, bool)

	// ResolveS3 is an S3 archive with its credentials, decrypted for the engine:
	// the cloud storage of the backup repository or the one storage names, with
	// storage's own key in place of the storage's when it has one, and its
	// prefix. Settings are read from scope. A volume repository has no
	// credentials, and is refused as ERR_STORAGE_TYPE_UNSUPPORTED.
	ResolveS3(ctx context.Context, db database.IDB, scope *entity.ObjectScope, storage entity.ManagedStorage) (
		*managedmodel.S3Archive, error)

	// MarkRestorePoint makes a restore point on a managed database (design §4,
	// "Restore points"): its engine marks it and waits until the WAL it is in is
	// archived. It runs in a transaction of its own, which holds the database's
	// row while the mark runs - never inside a longer one, such as a deployment's.
	MarkRestorePoint(ctx context.Context, req *MarkPointReq) (*MarkedPoint, error)

	// PruneDeployPoints removes the points deployments made that no base backup
	// kept reaches any more, given a nightly run's outputs, and answers their
	// names. A person's points are never removed.
	PruneDeployPoints(ctx context.Context, db database.IDB, projectID, appID string, outputs map[string]string) (
		[]string, error)
	// AllowReplicas runs the engine's command that lets replicas into a database,
	// in its task, the replication password on its stdin: never an argument, never
	// the environment.
	AllowReplicas(ctx context.Context, db database.Tx, app *entity.App, allow *entity.CommandTemplate,
		password string) error
	// WriterOf is the cluster's primary the engine's commands run on: the app, or
	// the member promoted over it.
	WriterOf(ctx context.Context, db database.IDB, app *entity.App, managed *entity.AppManagedSettings) (
		*entity.App, error)
	// RunCommand runs one of the engine's commands in the app's task and answers
	// its exit code: nil when it did not start, the app running no container.
	RunCommand(ctx context.Context, db database.Tx, app *entity.App, command *entity.CommandTemplate) (*int, error)
	// StopAndWait stops the app's service and waits, up to wait, until none of
	// its tasks runs: false when that was not seen in time - a task on a node
	// that does not answer is never seen stopped.
	StopAndWait(ctx context.Context, app *entity.App, wait time.Duration) (bool, error)
	// Start starts the app's service again, at the replicas it was stopped at.
	Start(ctx context.Context, app *entity.App) error
}

// MarkPointReq asks for a restore point on a managed database.
type MarkPointReq struct {
	ProjectID string
	AppID     string
	Name      string
	Note      string
	KeepUntil time.Time
	// By is the user who asks, empty for a deployment.
	By     string
	Origin string
	// ExistingOK answers a point of that name already there as made: a
	// deployment that is run again asks for the one it made the first time.
	ExistingOK bool
	// Record runs in the point's transaction once it is saved, for the caller's
	// audit; nil records nothing.
	Record func(ctx context.Context, db database.IDB, app *entity.App, point *entity.ManagedRestorePoint) error
}

// MarkedPoint is a point made, or found when ExistingOK asked for it.
type MarkedPoint struct {
	App     *entity.App
	Point   *entity.ManagedRestorePoint
	Existed bool
}
