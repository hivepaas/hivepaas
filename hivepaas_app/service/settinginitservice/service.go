package settinginitservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	InitDefaults(ctx context.Context, db database.IDB) error
	InitDefaultsWithTx(ctx context.Context, db database.Tx) error

	// InitSelfSignedCert creates the certificate an app is served with until a
	// real one exists for its domain. It belongs to installing HivePaaS rather
	// than to the defaults above: those are settings that have to exist for a
	// screen to open, and are filled in whenever one is found missing, while a
	// certificate is a thing an operator owns - one they removed on purpose must
	// stay removed.
	InitSelfSignedCert(ctx context.Context, db database.Tx) error

	// MoveDailyJobs moves the system jobs that run daily - the cleanup, the
	// backup, the certificates' renewal, the backup repositories' cleanup - to
	// their time of day in to, each whose schedule is still the one HivePaaS
	// gave it in from. One an administrator changed is left as it is. It answers
	// the jobs moved, for the caller to schedule again, and the names of those
	// left.
	MoveDailyJobs(ctx context.Context, db database.Tx, from, to *time.Location) (
		moved []*entity.Setting, left []string, err error)
}
