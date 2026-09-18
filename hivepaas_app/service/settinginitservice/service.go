package settinginitservice

import (
	"context"

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
}
