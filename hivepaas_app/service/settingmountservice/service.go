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
	// ClaimedPaths are the paths the app's active secrets, config files and
	// entries give files to, each with what gives it, leaving out the setting
	// exceptSettingID - the one being saved.
	ClaimedPaths(ctx context.Context, db database.IDB, appID, exceptSettingID string) (map[string]string, error)
	// EntryStates says, for each of the app's entries by setting id, what of it
	// is mounted, and why nothing is when nothing is.
	EntryStates(ctx context.Context, db database.IDB, app *entity.App) (map[string]*EntryState, error)

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

// Why nothing of an entry is mounted.
const (
	ReasonEntryDisabled     = "entry-disabled"
	ReasonKeyInvalid        = "key-invalid"
	ReasonSourceUnavailable = "source-unavailable" // missing, disabled, or not visible from the app
	ReasonSourceIncomplete  = "source-incomplete"  // a required part is empty: a certificate not obtained yet
	ReasonPathsTaken        = "paths-taken"        // every path is another entry's
)

// EntryState is what of an entry is mounted.
type EntryState struct {
	// Reason is why nothing is mounted; empty when something is.
	Reason  string   `json:"reason,omitempty"`
	Mounted []string `json:"mounted"`
}
