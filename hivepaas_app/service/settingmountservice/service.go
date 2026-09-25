package settingmountservice

import (
	"context"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

type Service interface {
	// Resolve is the files app should have now: those of its active entries
	// whose source is active, visible from its scope and usable.
	Resolve(ctx context.Context, db database.IDB, app *entity.App) ([]*File, error)
	// ApplyToService brings spec's mounted files to what Resolve says, creating
	// the objects Docker does not have yet. It is called inside the caller's
	// service update, and again on each retry of it.
	ApplyToService(ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec) error
	// Sweep removes the app's mounted objects its service no longer references.
	// It is called once the update is made; what stays in use is left for the
	// next sweep.
	Sweep(ctx context.Context, app *entity.App) error
	// Refresh is one service update that applies, then a sweep.
	Refresh(ctx context.Context, db database.IDB, app *entity.App) error
	// RemoveApp removes every mounted object of an app whose service is gone.
	RemoveApp(ctx context.Context, appID string) error

	// RecordRefresh records, in db's transaction, a refresh of the apps that
	// read one of settings: an entry's own app, or the apps whose entries mount
	// a source. It records nothing, and returns nil, when no app does.
	// The caller hands the task to the queue once its transaction has committed;
	// this service does not hold the queue, which depends on the app service.
	RecordRefresh(ctx context.Context, db database.IDB, settings ...*entity.Setting) (*entity.Task, error)
}

// File is one part of a source, at a path of an app's containers.
type File struct {
	// Entry is the entry's key, or TLSEntry.
	Entry     string
	Part      string
	Path      string
	UID       string
	GID       string
	Mode      fileutil.FileMode
	Sensitive bool
	Data      []byte
	// Rotation is the part's RotationKey: a new one is a new object.
	Rotation string
}
